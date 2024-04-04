package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"slices"
	"strings"
	"syscall"

	"github.com/jchv/go-webview2"
)

const appName = "oasiz-terminal"

func main() {
	// キャッシュディレクトリを作成
	cacheDir, err := createCacheDir()
	if err != nil {
		panic(err)
	}

	// ttyd をインストール
	ttydExecutablePath, err := installTtyd(ttydDownloadUrl, cacheDir, ttydExecutableName)
	if err != nil {
		panic(err)
	}

	// 空きポートの確認
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	listener.Close()

	// ttyd を起動
	_, port, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		panic(err)
	}
	command := []string{"powershell.exe"}
	if len(os.Args) > 1 {
		command = os.Args[1:]
	}
	ttyExec := startTtyd(ttydExecutablePath, port, command)
	defer ttyExec.Process.Kill()

	// ブラウザを起動
	w := webview2.NewWithOptions(webview2.WebViewOptions{
		Debug:     true,
		AutoFocus: true,
		WindowOptions: webview2.WindowOptions{
			Title:  "OASIZ Terminal",
			Width:  800,
			Height: 600,
			IconId: 2, // icon resource id
			Center: true,
		},
	})
	if w == nil {
		log.Fatalln("Failed to load webview.")
	}
	defer w.Destroy()
	w.SetSize(800, 600, webview2.HintNone)
	w.Navigate("http://127.0.0.1:" + port)
	w.Run()

}

func startTtyd(ttydExecutablePath string, port string, command []string) *exec.Cmd {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	ttydArgs := []string{
		"-p", port,
		"--writable", "--once",
		"-t", "enableSixel=true",
		"-t", "disableReconnect=true",
		"winpty"}

	ttydArgs = slices.Concat(ttydArgs, command)

	fmt.Printf("Start ttyd: `%s \"%s\"`\n", ttydExecutablePath, strings.Join(ttydArgs, "\" \""))
	ttydExec := exec.CommandContext(ctx, ttydExecutablePath, ttydArgs...)
	ttydExec.Stdout = os.Stdout
	ttydExec.Stderr = os.Stderr
	ttydExec.Cancel = func() error {
		fmt.Fprintf(os.Stderr, "Receive SIGINT.\n")
		return ttydExec.Process.Signal(os.Interrupt)
	}
	ttydExec.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000000}

	err := ttydExec.Start()
	if err != nil {
		panic(err)
	}

	return ttydExec
}


// ttyd を格納するディレクトリを作成する。
// 作成したディレクトリのパスを返却する。
func createCacheDir() (string, error) {
	var baseDir, err = os.UserCacheDir()
	if err != nil {
		return "", err
	}
	var cacheDir = filepath.Join(baseDir, appName)
	if err := os.MkdirAll(cacheDir, 0766); err != nil {
		return "", err
	}
	return cacheDir, nil
}

// 単純なファイル配置でインストールが完了するもののインストール処理。
//
// downloadUrl からファイルをダウンロードし、 installDir に fileName とう名前で配置する。
func installTtyd(downloadUrl string, installDir string, fileName string) (string, error) {

	// ツールの配置先組み立て
	filePath := filepath.Join(installDir, fileName)

	// ダウンロード
	err := download(downloadUrl, filePath)
	if err != nil {
		return filePath, err
	}

	// 実行権限の付与
	err = addExecutePermission(filePath)
	if err != nil {
		return filePath, err
	}

	// 互換性モードの設定
	fmt.Println("reg", "add", `HKCU\Software\Microsoft\Windows NT\CurrentVersion\AppCompatFlags\Layers`, "/v", filePath, "/d", "WIN8RTM", "/f")
	regExec := exec.Command("reg", "add", `HKCU\Software\Microsoft\Windows NT\CurrentVersion\AppCompatFlags\Layers`, "/v", filePath, "/d", "WIN8RTM", "/f")
	fmt.Println(regExec)

	err = regExec.Run()
	if err != nil {
		return filePath, err
	}

	return filePath, nil
}

// ファイルダウンロード処理。
//
// downloadUrl からファイルをダウンロードし、 destPath へ配置する。
func download(downloadUrl string, destPath string) error {
	if isExists(destPath) {
		fmt.Printf("%s aleady exist, use this.\n", filepath.Base(destPath))
	} else {
		fmt.Printf("Download %s from %s ...", filepath.Base(destPath), downloadUrl)

		// HTTP GETリクエストを送信
		resp, err := http.Get(downloadUrl)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		// ファイルを作成
		out, err := os.Create(destPath)
		if err != nil {
			return err
		}
		defer out.Close()

		// レスポンスの内容をファイルに書き込み
		_, err = io.Copy(out, resp.Body)
		if err != nil {
			return err
		}

		fmt.Printf(" done.\n")
	}

	return nil
}

func addExecutePermission(filePath string) error {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return err
	}
	fileMode := fileInfo.Mode()
	err = os.Chmod(filePath, fileMode|0111)
	if err != nil {
		return err
	}

	return nil
}

func isExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}
