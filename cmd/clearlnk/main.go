//go:build windows && clearlnk
// +build windows,clearlnk

package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	lnk "github.com/parsiya/golnk"
)

func main() {
	log.SetFlags(0)

	dirs := knownShortcutDirs()
	if len(dirs) == 0 {
		log.Println("未找到可扫描的快捷方式目录")
		return
	}

	invalid := make([]invalidShortcut, 0)
	for _, dir := range dirs {
		paths, err := findInvalidShortcuts(dir)
		if err != nil {
			log.Printf("扫描目录 %s 失败: %v\n", dir, err)
			continue
		}
		invalid = append(invalid, paths...)
	}

	for _, s := range invalid {
		if s.reason != "" {
			fmt.Printf("%s: %s\n", s.path, s.reason)
		} else {
			fmt.Println(s.path)
		}
	}
}

type invalidShortcut struct {
	path   string
	reason string
}

// knownShortcutDirs returns existing Windows shortcut directories we should scan.
func knownShortcutDirs() []string {
	userProfile := os.Getenv("USERPROFILE")
	public := os.Getenv("PUBLIC")
	appData := os.Getenv("APPDATA") // Roaming
	programData := os.Getenv("PROGRAMDATA")

	candidates := []string{
		filepath.Join(userProfile, "Desktop"),
		filepath.Join(public, "Desktop"),
		filepath.Join(appData, "Microsoft", "Windows", "Start Menu"),
		filepath.Join(programData, "Microsoft", "Windows", "Start Menu"),
		filepath.Join(appData, "Microsoft", "Internet Explorer", "Quick Launch"),
		filepath.Join(appData, "Microsoft", "Internet Explorer", "Quick Launch", "User Pinned", "TaskBar"),
		filepath.Join(appData, "Microsoft", "Internet Explorer", "Quick Launch", "User Pinned", "StartMenu"),
	}

	seen := make(map[string]struct{})
	existing := make([]string, 0, len(candidates))

	for _, dir := range candidates {
		if dir == "" {
			continue
		}
		if _, err := os.Stat(dir); err != nil {
			continue
		}
		if _, ok := seen[dir]; ok {
			continue
		}
		seen[dir] = struct{}{}
		existing = append(existing, dir)
	}

	return existing
}

func findInvalidShortcuts(root string) ([]invalidShortcut, error) {
	invalid := make([]invalidShortcut, 0)
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if !strings.EqualFold(filepath.Ext(path), ".lnk") {
			return nil
		}

		ok, reason := shortcutValid(path)
		if !ok {
			invalid = append(invalid, invalidShortcut{path: path, reason: reason})
		}
		return nil
	})
	return invalid, err
}

// shortcutValid determines whether the shortcut points to an existing target.
func shortcutValid(path string) (bool, string) {
	target, reason := resolveShortcutTarget(path)
	if target == "" {
		if reason == "" {
			reason = "未找到目标路径"
		}
		return false, reason
	}

	if _, err := os.Stat(target); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, "目标不存在: " + target
		}
		return false, fmt.Sprintf("读取目标失败: %v", err)
	}

	return true, ""
}

// parseLnk tries the normal parser first, then falls back to a best-effort
// parser that ignores ExtraData parsing errors. Some lnks have truncated
// ExtraData blocks; we still want the target path from earlier sections.
func parseLnk(path string) (*lnk.LnkFile, error) {
	link, err := lnk.File(path)
	if err == nil {
		return &link, nil
	}
	if bestEffort, fallbackErr := parseLnkBestEffort(path); fallbackErr == nil {
		return bestEffort, nil
	}
	return nil, err
}

// parseLnkBestEffort replicates lnk.File but skips ExtraData parsing.
// It returns enough fields (LinkInfo, StringData) for target resolution.
func parseLnkBestEffort(filename string) (*lnk.LnkFile, error) {
	fi, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fi.Close()

	maxSize := uint64(1 << 22)
	if s, err := fi.Stat(); err == nil {
		if maxSize > 0 {
			maxSize = uint64(s.Size())
		}
	}

	var f lnk.LnkFile

	f.Header, err = lnk.Header(fi, maxSize)
	if err != nil {
		return nil, err
	}

	if f.Header.LinkFlags["HasLinkTargetIDList"] {
		f.IDList, err = lnk.LinkTarget(fi)
		if err != nil {
			return nil, err
		}
	}

	if f.Header.LinkFlags["HasLinkInfo"] {
		f.LinkInfo, err = lnk.LinkInfo(fi, maxSize)
		if err != nil {
			return nil, err
		}
	}

	f.StringData, err = lnk.StringData(fi, f.Header.LinkFlags)
	if err != nil {
		return nil, err
	}

	// We intentionally skip ExtraData blocks here.
	return &f, nil
}

// shortcutTarget extracts and normalizes the target path from a ShellLink file.
func shortcutTarget(linkPath string, link *lnk.LnkFile) string {
	path := ""

	base := strings.TrimSpace(link.LinkInfo.LocalBasePath)
	if base == "" {
		base = strings.TrimSpace(link.LinkInfo.LocalBasePathUnicode)
	}

	network := strings.TrimSpace(link.LinkInfo.NetworkRelativeLink.NetName)
	if network == "" {
		network = strings.TrimSpace(link.LinkInfo.NetworkRelativeLink.NetNameUnicode)
	}

	suffix := strings.TrimSpace(link.LinkInfo.CommonPathSuffix)
	if suffix == "" {
		suffix = strings.TrimSpace(link.LinkInfo.CommonPathSuffixUnicode)
	}

	switch {
	case base != "" && suffix != "":
		path = filepath.Join(base, suffix)
	case base != "":
		path = base
	case network != "" && suffix != "":
		path = filepath.Join(network, suffix)
	case network != "":
		path = network
	case suffix != "":
		path = suffix
	}

	if path == "" {
		if rel := strings.TrimSpace(link.StringData.RelativePath); rel != "" {
			path = rel
		} else if wd := strings.TrimSpace(link.StringData.WorkingDir); wd != "" {
			path = wd
		}
	}

	if path == "" {
		return ""
	}

	path = strings.Trim(path, "\"")
	path = expandEnv(path)

	// UNC paths are absolute but filepath.IsAbs already handles them on Windows.
	if !filepath.IsAbs(path) {
		path = filepath.Join(filepath.Dir(linkPath), path)
	}

	return filepath.Clean(path)
}

// expandEnv expands both %VAR% (Windows style) and $VAR (Go style) environment placeholders.
func expandEnv(path string) string {
	if path == "" {
		return ""
	}

	// Replace %VAR% placeholders first.
	percentEnv := regexp.MustCompile(`%([^%]+)%`)
	path = percentEnv.ReplaceAllStringFunc(path, func(s string) string {
		key := s[1 : len(s)-1]
		if key == "" {
			return s
		}
		if val, ok := os.LookupEnv(key); ok {
			return val
		}
		// If variable is unknown, keep the original token so caller can see it.
		return s
	})

	// Then let os.ExpandEnv handle $VAR / ${VAR} style placeholders.
	return os.ExpandEnv(path)
}

var (
	oleInitOnce sync.Once
	oleInitErr  error
)

// resolveShortcutTargetCOM uses WScript.Shell COM APIs to resolve .lnk target paths.
func resolveShortcutTargetCOM(path string) (string, error) {
	oleInitOnce.Do(func() {
		oleInitErr = ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED)
	})
	if oleInitErr != nil {
		return "", oleInitErr
	}

	shellObj, err := oleutil.CreateObject("WScript.Shell")
	if err != nil {
		return "", err
	}
	defer shellObj.Release()

	shell, err := shellObj.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return "", err
	}
	defer shell.Release()

	shortcut, err := oleutil.CallMethod(shell, "CreateShortcut", path)
	if err != nil {
		return "", err
	}
	defer shortcut.ToIDispatch().Release()

	target, err := oleutil.GetProperty(shortcut.ToIDispatch(), "TargetPath")
	if err != nil {
		return "", err
	}
	targetPath := strings.TrimSpace(target.ToString())
	if targetPath == "" {
		return "", fmt.Errorf("COM 未返回目标路径")
	}

	return filepath.Clean(expandEnv(targetPath)), nil
}

// resolveShortcutTarget tries COM resolution first (Shell link target),
// then falls back to manual LNK parsing.
func resolveShortcutTarget(path string) (string, string) {
	if target, err := resolveShortcutTargetCOM(path); err == nil && target != "" {
		return target, ""
	}

	link, err := parseLnk(path)
	if err != nil {
		return "", fmt.Sprintf("无法解析快捷方式: %v", err)
	}

	target := shortcutTarget(path, link)
	if target == "" {
		return "", "未找到目标路径"
	}

	return target, ""
}
