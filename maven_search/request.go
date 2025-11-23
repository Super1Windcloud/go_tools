package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	_ "os"
	"time"
)
import "github.com/fatih/color"

var red = color.New(color.FgRed).Add(color.Bold)

func searchFromMaven(query string) error {

	if query == "" || len(query) == 0 {

		err := errors.New("请输入你要查询的包名")
		_, _ = red.Println(err)
		return err
	} else {
		packageResult, err := getPureResultsFromMaven(query)
		if err != nil {
			_, _ = red.Println(err)
			return err
		}
		displayResults(packageResult)
	}
	return nil
}

func getPureResultsFromMaven(query string) ([]Doc, error) {
	result, err := startSearch(query)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func displayResults(result []Doc) {
	green := color.New(color.FgGreen).Add(color.Bold)
	_ , _ =  green.Println("查询结果个数：", len(result) )
	_ ,_ = green.Println("groupId \t\t\t\t\t   artifactId\t\t\t\t\t      version")
	for _, doc := range result {
		fmt.Printf("%-50s %-50s %-20s\n",  doc.GroupId,
			 doc.ArtifactId, doc.Version)
	}
}

func startSearch(query string) ([]Doc, error) {
	result, err := fetchMavenSearchResults(query)
	if err != nil {
		return nil, err
	}
	// go 的重新赋值必须保证变量类型一致 ,而rust let 可以自动推断类型进行类型转换
	paserResult, err := parseMavenSearchJSONResults(result)
	if err != nil {
		return nil, err
	}
	return paserResult, nil
}

func parseMavenSearchJSONResults(response []string) ([]Doc, error) {
	// 解析html 数据

	var hits_result SearchResponse
	err := json.Unmarshal([]byte(response[0]), &hits_result)
	if err != nil {
		_, _ = red.Println("JSON  解析失败：", err)
		return nil, err
	}
	var docResult []Doc
	for _, doc := range hits_result.Response.Docs {
		docResult = append(docResult, doc)
	}
	if docResult == nil || len(docResult) == 0 {
		_, _ = red.Println("查询结果为空")
		err := errors.New("查询结果为空")
		return nil, err
	}
	return docResult, nil
}

// 定义与 JSON 响应结构对应的 Go 结构体
type Doc struct {
	GroupId string `json:"g"`
	ArtifactId string `json:"a"`
	Version  string `json:"latestVersion"`
	//ApacheMaven string `json:"p"`
	//GradleGroovy string `json:"gav"`
	//GradleKotlin string `json:"gav"`
	//ScalaSbt string `json:"gav"`
	//ApacheIvy string `json:"p"`
	//GroovyGrape string `json:"gav"`
	//Leiningen string `json:"gav"`
	//Bazel  string `json:"p"`
	//ApacheBuildr string `json:"p"`
    //PURL string `json:"p"`
}

type Response struct {
	Docs []Doc `json:"docs"`
}

type SearchResponse struct {
	Response Response `json:"response"`
}
/**
maven  deps
<dependency>
   <groupId>com.github.sixinyiyu</groupId>
   <artifactId>http-spring-boot-start</artifactId>
   <version>1.0.3.RELEASE</version>
</dependency>

gradle deps
implementation 'com.github.sixinyiyu:http-spring-boot-start:1.0.3.RELEASE'

*/

func fetchMavenSearchResults(searchTerm string, proxyURL ...string) ([]string, error) {
	// 创建 HTTP 客户端
	client := &http.Client{
		Timeout: 10 * time.Second, // 设置请求超时
	}
	if len(proxyURL) == 0 { // 如果没有代理，可以设置为 nil
		proxyURL = nil
	}
	// 构建请求 URL
	url := "https://search.maven.org/solrsearch/select?rows=100&q=" + searchTerm
	// 设置代理（如果有）
	proxy := http.ProxyURL(nil) // 如果没有代理，可以设置为 nil
	client.Transport = &http.Transport{
		Proxy: proxy,
	}

	// 发送 GET 请求
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %v", err)
	}

	// 调用处理函数，将响应传递给回调函数
	str := string(body)
	result := append([]string{}, str)
	return result, nil
}
func SearchFromMavenToMavenDeps(query string) error {
	red := color.New(color.FgRed).Add(color.Bold)
	green := color.New(color.FgGreen).Add(color.Bold)
	if query == "" || len(query) == 0 {
		err := errors.New("请输入你要查询的包名")
		_, _ = red.Println(err)
		return err
	} else {
		 mavenPackageResult, err := getPureResultsFromMaven(query)
		if err != nil {
			_, _ = red.Println(err)
			return err
		}
		var groupIDArr  []string
		var artifactIdArr  []string
		var versionArr  []string
		 var   homepagesUrl  []string
		for _, doc := range mavenPackageResult {
			groupIDArr = append(groupIDArr, doc.GroupId)
			artifactIdArr = append(artifactIdArr, doc.ArtifactId)
			versionArr = append(versionArr, doc.Version)
	  // 	https://search.maven.org/artifact/com.github.sixinyiyu/http-spring-boot-start/1.0.3.RELEASE/jar
	        homepagesUrl = append(homepagesUrl,
				fmt.Sprintf("https://search.maven.org/artifact/%s/%s/%s/jar", doc.GroupId, doc.ArtifactId, doc.Version))
		}
	currentItem :=	fancyMavenList( groupIDArr, artifactIdArr, versionArr, homepagesUrl)
		templateString  := fmt.Sprintf(`	
			 <dependency>
				 <groupId>%s</groupId>
  	   		  	 <artifactId>%s</artifactId>
  	   	  		 <version>%s</version>
			 </dependency>
			 `, currentItem.groupID, currentItem.artifactId, currentItem.version)
		_ ,_  = green.Println( templateString )
	}
	return nil
}

func getResultsWithDescriptionFromMaven(query string) ( []Doc, error) {
	result, err := fetchWholeSearchResults(query)
	if err != nil {
		return nil, err
	}
	// go 的重新赋值必须保证变量类型一致 ,而rust let 可以自动推断类型进行类型转换
	paserResult, err := parseMavenSearchHTMLResults(result)
	if err != nil {
		return nil, err
	}
	return paserResult, nil


}

func parseMavenSearchHTMLResults(html  []string) ([]Doc, error ) {
//	var docResult []Doc
//	  //  goquery 解析html 数据
//	  doc, err := goquery.NewDocumentFromReader(strings.NewReader(html[0]))
//if err != nil {
//		 return nil, err
//	}
	//fmt.Println(doc.Html())
	//  os.Exit(0)
	//doc.Find(".search-result").Each(func(i int, s *goquery.Selection) {
	//	// 解析每个搜索结果的信息
	//	groupId := s.Find(".artifact-id").Text()
	//	artifactId := s.Find(".artifact-id").Next().Text()
	//	version := s.Find(".version").Text()
	//	description := s.Find(".description").Text()
	//	docResult = append(docResult, Doc{GroupId: groupId, ArtifactId: artifactId, Version: version, ApacheMaven: description})
	//})
	//if docResult == nil || len(docResult) == 0 {
	//	err := errors.New("查询结果为空")
	//	return nil, err
	//}
	//fmt.Println(docResult)

	return  nil , nil
}

func fetchWholeSearchResults(query string) (  []string  , error ) {

    result, err := getPureResultsFromMaven(query)
	var currentGroupId string
	var currentArtifactId string
	var currentVersion string
	for   _, doc := range result {
		if  query==doc.ArtifactId {

	     currentGroupId = doc.GroupId
			currentArtifactId = doc.ArtifactId
			currentVersion = doc.Version
		 }
	}
	 //"https://search.maven.org/artifact/com.github.sixinyiyu/http-spring-boot-start/1.0.3.RELEASE/jar"
	var  url string
	for _, doc := range result {
      if doc.GroupId==currentGroupId && doc.ArtifactId == currentArtifactId && doc.Version == currentVersion {
		   url = fmt.Sprintf("https://search.maven.org/artifact/%s/%s/%s/jar", doc.GroupId, doc.ArtifactId, doc.Version)
	  }
	}
     fmt.Println(url)
	client := &http.Client{
		Timeout: 10 * time.Second, // 设置请求超时
	}

	// 构建请求 URL

	// 设置代理（如果有）
	proxy := http.ProxyURL(nil) // 如果没有代理，可以设置为 nil
	client.Transport = &http.Transport{
		Proxy: proxy,
	}

	// 发送 GET 请求
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %v", err)
	}

	// 调用处理函数，将响应传递给回调函数
	str := string(body)
	resultFinal  := append([]string{}, str)
	return resultFinal, nil
}

func searchFromMavenToGradleDeps(query string) error {
	red := color.New(color.FgRed).Add(color.Bold)
	green := color.New(color.FgGreen).Add(color.Bold)
	if query == "" || len(query) == 0 {
		err := errors.New("请输入你要查询的包名")
		_, _ = red.Println(err)
		return err
	} else {
		mavenPackageResult, err := getPureResultsFromMaven(query)
		if err != nil {
			_, _ = red.Println(err)
			return err
		}
		var groupIDArr  []string
		var artifactIdArr  []string
		var versionArr  []string
		var   homepagesUrl  []string
		for _, doc := range mavenPackageResult {
			groupIDArr = append(groupIDArr, doc.GroupId)
			artifactIdArr = append(artifactIdArr, doc.ArtifactId)
			versionArr = append(versionArr, doc.Version)
			// 	https://search.maven.org/artifact/com.github.sixinyiyu/http-spring-boot-start/1.0.3.RELEASE/jar
			homepagesUrl = append(homepagesUrl,
				fmt.Sprintf("https://search.maven.org/artifact/%s/%s/%s/jar", doc.GroupId, doc.ArtifactId, doc.Version))
		}
		currentItem :=	fancyMavenList( groupIDArr, artifactIdArr, versionArr, homepagesUrl)
	 _ , _ = 	red.Println( "Gradle Kotlin format ")
	 //implementation("com.github.sixinyiyu:http-spring-boot-start:1.0.3.RELEASE")
		templateString  := fmt.Sprintf(
			 `implementation("%s:%s:%s")`,currentItem.groupID, currentItem.artifactId, currentItem.version)
		_ ,_  = green.Println( templateString )

		_ , _ = 	red.Println( "Gradle Groovy format ")
		templateString  = fmt.Sprintf(`implmentation '%s:%s:%s'`, currentItem.groupID, currentItem.artifactId, currentItem.version)
		_ ,_  = green.Println( templateString )

	}
	return nil

}
