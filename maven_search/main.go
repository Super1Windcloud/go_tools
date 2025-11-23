package main

import (
	"fmt"
	"log"
	"os"

	"github.com/fatih/color"
	cli "github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:    "maven_search",
		Version: "1.0.0",
		Authors: []*cli.Author{
			{
				Name:  "Superwindcloud",
				Email: "EMAIL			 https://gitee.com/Superwindcloud",
			},
		},

		Commands: []*cli.Command{
			{
				Name:      "search",
				Usage:     "查询maven仓库中的包信息",
				UsageText: "maven_search s   spring-boot , maven_search s  -m spring , maven_search s -g spring ",
				HelpName:  "search",
				Aliases:   []string{"s"},
				ArgsUsage: "maven_search search   spring-boot ",
				Action: func(args *cli.Context) error {
					if args.IsSet("m") {
						// 获取 -m Value 值
						mavenQuery := args.String("m") ;
						err  := SearchFromMavenToMavenDeps(mavenQuery);
						if err!= nil {
							return err
						}
						return  nil
					}else if args.IsSet("g") {
						// 获取 -g Value 值
						gradleQuery := args.String("g") ;
						err  := searchFromMavenToGradleDeps(gradleQuery);
						if err!= nil {
							return err
						}
						return  nil
					}
					if args.NArg() == 0 {
						err := cli.ShowSubcommandHelp(args)
						if err != nil {
							return err
						}
						return nil
					} else  {
						query := args.Args().Get(0)
						err := searchFromMaven(query )
						if err != nil {
							 return fmt.Errorf("failed to search package: %w", err)
						}
					}
					return nil
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "m",
						Aliases: []string{"maven"},
						Usage:   "以Maven依赖形式输出",
					},
					&cli.StringFlag{
						Name:    "g",
						Aliases: []string{"gradle"},
						Usage:   "以Gradle依赖形式输出",
					},
				},
			},
		},
		Usage:           "命令行工具，用于查询maven仓库中的包信息",
		HideHelpCommand: true,
		Action: func(args *cli.Context) error {
			if args.NArg() == 0 || args.Args().Len() == 0 {
				err := cli.ShowAppHelp(args)
				if err != nil {
					return fmt.Errorf("failed to open file: %w", err)
				}
				return nil
			}
			if args.Command == nil {
				_, _ =  color.New(color.FgRed).Add(color.Bold).Println("未知命令: %s", args.Args().Get(0))
				return fmt.Errorf("未知命令: %s", args.Args().Get(0))
			}
			return nil
		},
		OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
			if isSubcommand {
				return cli.Exit("未知命令: "+c.Command.Name, 1)
			}
			return cli.Exit("命令使用错误: "+err.Error(), 1)
		},

	}
	if err := app.Run(os.Args); err != nil {

		log.Fatal(err)
	}


}
