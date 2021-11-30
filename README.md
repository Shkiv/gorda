# gorda
Client for [gotracked](https://github.com/Shkiv/gotracked)

## Building and running
### Linux
1. Install [gotracked](https://github.com/Shkiv/gotracked)
2. Install GTK, GLib, Cairo
4. Use `go run .` and `go build .` to run or build
5. (optional) Use Glade for edit the interface

### Windows
(one of the possible ways)
1. Install [gotracked](https://github.com/Shkiv/gotracked)
2. Install dependencies in the MinGW:
```sh
pacman -S mingw-w64-x86_64-gtk3 mingw-w64-x86_64-toolchain base-devel glib2-devel
```
3. Patch `pkgconfig`:
```
sed -i -e 's/-Wl,-luuid/-luuid/g' /mingw64/lib/pkgconfig/gdk-3.0.pc
```
4. Use `go run .` and `go build .` from PowerShell or MinGW
5. (optional) Install Glade:
```
pacman -S mingw-w64-x86_64-glade
```
6. (optional) Run Glade from PowerShell or MinGW for edit the interface