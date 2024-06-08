package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"

	webview "github.com/webview/webview_go"
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
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	_, port, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		panic(err)
	}

	// ブラウザを起動(表示はまだ)
	debug := true
	w := webview.New(debug)
	defer w.Destroy()
	w.SetTitle("OASIZ Terminal")
	w.SetSize(960, 640, webview.HintNone)

	// ttyd の開始情報組み立て
	ttydArgs := []string{"--writable", "--port", port, "--once", "-t", "enableSixel=true", os.Args[1]}
	fmt.Printf("Start ttyd: `%s \"%s\"`\n", ttydExecutablePath, strings.Join(ttydArgs, "\" \""))
	ttydExec := exec.CommandContext(ctx, ttydExecutablePath, ttydArgs...)
	ttydExec.Stdout = os.Stdout
	ttydExec.Stderr = os.Stderr

	// Interrupt で、ブラウザを閉じて ttyd を殺して終了する
	ttydExec.Cancel = func() error {
		fmt.Fprintf(os.Stderr, "Receive SIGINT.\n")
		w.Terminate()
		ttydExec.Process.Kill()
		os.Exit(0)
		return nil
	}

	// ttyd 起動
	err = ttydExec.Start()
	if err != nil {
		panic(err)
	}

	// ブラウザ表示
	fmt.Println("Open browser: http://localhost:" + port)
	w.Navigate("http://localhost:" + port)
	w.Run()

	ttydExec.Process.Kill()
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
