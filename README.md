# oasiz-terminal

## Usage:

```sh
# Open terminal and run bash
oasiz-terminal bash
```

## Start dev environment:

```sh
devcontainer.vim-linux-amd64 start .
```

See: [mikoto2000/devcontainer.vim: コンテナ上で Vim を使った開発をするためのツール。](https://github.com/mikoto2000/devcontainer.vim)

### memo

```sh
apt update

# ttyd を Linux でビルドするための依存関係
apt install -y build-essential cmake git libjson-c-dev libwebsockets-dev


# Linux で webkit_go を利用するのに必要な依存関係
apt install -y libgtk-3-dev libwebkit2gtk-4.0-dev
```

## Build:

Windows:

`/go/pkg/mod/github.com/webview/webview_go@<バージョン>/libs/webview/include/EventToken.h` に格納。

※ `EventToken.h` は [WinLibs - GCC+MinGW-w64 compiler for Windows](https://winlibs.com/) から取得。

ビルドに必要なパッケージは以下。

```sh
sudo apt install gcc-multilib gcc-mingw-w64 g++-mingw-w64 binutils-mingw-w64
```

以下コマンドでビルド。

```sh
CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ GOOS=windows go build
```

## License:

Copyright (C) 2024 mikoto2000

This software is released under the MIT License, see LICENSE

このソフトウェアは MIT ライセンスの下で公開されています。 LICENSE を参照してください。


## Author:

mikoto2000 <mikoto2000@gmail.com>


## 参考資料

- [fatal error: EventToken.h: No such file or directory · Issue #1036 · webview/webview](https://github.com/webview/webview/issues/1036)

