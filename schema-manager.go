package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/spf13/cobra"
)

var (
	repoURL    = "https://github.com/opencommand/commands"
	cacheDir   string
	forceClone bool
)

func main() {
	// 获取用户主目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error getting user home directory: %v\n", err)
		os.Exit(1)
	}
	cacheDir = filepath.Join(homeDir, ".opencmd", "commands")

	var rootCmd = &cobra.Command{
		Use:   "schema-manager",
		Short: "A tool to manage command schemas from GitHub repository",
		Long:  `Schema Manager is a CLI tool for managing command schemas from the opencommand/commands repository.`,
	}

	var initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize by cloning the repository to cache directory",
		Long:  `Clone the opencommand/commands repository to the user's cache directory.`,
		Run: func(cmd *cobra.Command, args []string) {
			initRepository()
		},
	}

	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "List all .hl files in the cache directory",
		Long:  `List all .hl files in the cache directory organized by directory tree.`,
		Run: func(cmd *cobra.Command, args []string) {
			listFiles()
		},
	}

	var searchCmd = &cobra.Command{
		Use:   "search [pattern]",
		Short: "Search for .hl files matching a pattern",
		Long:  `Search for .hl files in the cache directory using regex pattern.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			searchFiles(args[0])
		},
	}

	var statusCmd = &cobra.Command{
		Use:   "status",
		Short: "Check repository status and sync with remote",
		Long:  `Check if the local cached repository is synchronized with the remote repository.`,
		Run: func(cmd *cobra.Command, args []string) {
			checkRepository()
		},
	}

	// 添加标志
	initCmd.Flags().BoolVarP(&forceClone, "force", "f", false, "Force re-clone by removing existing cache")

	// 添加子命令
	rootCmd.AddCommand(initCmd, listCmd, searchCmd, statusCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initRepository() {
	// 如果强制克隆，先删除现有目录
	if forceClone {
		if err := os.RemoveAll(cacheDir); err != nil {
			fmt.Printf("Error removing existing directory: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Removed existing cache directory.")
	}

	// 检查目录是否已存在
	if _, err := os.Stat(cacheDir); err == nil && !forceClone {
		fmt.Printf("Repository already exists at: %s\n", cacheDir)
		fmt.Println("Use -f flag to force re-clone.")
		return
	}

	// 创建目录
	err := os.MkdirAll(cacheDir, 0755)
	if err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		os.Exit(1)
	}

	// 克隆仓库
	fmt.Printf("Cloning repository to: %s\n", cacheDir)
	_, err = git.PlainClone(cacheDir, &git.CloneOptions{
		URL: repoURL,
	})

	if err != nil {
		fmt.Printf("Error cloning repository: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Repository cloned successfully!")
}

func listFiles() {
	if !repositoryExists() {
		fmt.Println("Repository not found. Run 'schema-manager init' first.")
		return
	}

	fmt.Println("Listing .hl files in cache directory:")
	fmt.Println("=====================================")

	err := filepath.Walk(cacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".hl") {
			relPath, _ := filepath.Rel(cacheDir, path)
			fmt.Printf("  %s\n", relPath)
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
	}
}

func searchFiles(pattern string) {
	if !repositoryExists() {
		fmt.Println("Repository not found. Run 'schema-manager init' first.")
		return
	}

	regex, err := regexp.Compile(pattern)
	if err != nil {
		fmt.Printf("Invalid regex pattern: %v\n", err)
		return
	}

	fmt.Printf("Searching for .hl files matching pattern: %s\n", pattern)
	fmt.Println("==================================================")

	found := false
	err = filepath.Walk(cacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".hl") {
			// 只搜索文件名部分
			if regex.MatchString(info.Name()) {
				relPath, _ := filepath.Rel(cacheDir, path)
				fmt.Printf("  %s\n", relPath)
				found = true
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		return
	}

	if !found {
		fmt.Println("No .hl files found matching the pattern.")
	}
}

func checkRepository() {
	if !repositoryExists() {
		fmt.Println("Repository not found. Run 'schema-manager init' first.")
		return
	}

	// 打开仓库
	repo, err := git.PlainOpen(cacheDir)
	if err != nil {
		fmt.Printf("Error opening repository: %v\n", err)
		return
	}

	// 获取远程引用
	remote, err := repo.Remote("origin")
	if err != nil {
		fmt.Printf("Error getting remote: %v\n", err)
		return
	}

	// 获取远程分支信息
	refs, err := remote.List(&git.ListOptions{})
	if err != nil {
		fmt.Printf("Error listing remote refs: %v\n", err)
		return
	}

	// 获取本地HEAD
	head, err := repo.Head()
	if err != nil {
		fmt.Printf("Error getting HEAD: %v\n", err)
		return
	}

	// 查找远程main分支
	var remoteMainHash plumbing.Hash
	for _, ref := range refs {
		if ref.Name().IsBranch() && ref.Name().Short() == "main" {
			remoteMainHash = ref.Hash()
			break
		}
	}

	if remoteMainHash.IsZero() {
		fmt.Println("Could not find remote main branch.")
		return
	}

	// 比较本地和远程
	if head.Hash() == remoteMainHash {
		fmt.Println("✓ Local repository is up to date with remote.")
	} else {
		fmt.Println("✗ Local repository is behind remote.")
		fmt.Printf("  Local HEAD:  %s\n", head.Hash().String()[:8])
		fmt.Printf("  Remote main: %s\n", remoteMainHash.String()[:8])
		fmt.Println("  Run 'schema-manager init -f' to update.")
	}
}

func repositoryExists() bool {
	_, err := os.Stat(cacheDir)
	return err == nil
}
