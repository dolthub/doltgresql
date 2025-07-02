# Finding Extension Function Imports
These are commands that can be used to find the functions that an extension imports, so that we know which ones we need to implement for the extension to load.
## Windows
On Windows, we make use of `dumpbin`, which is installed alongside Visual Studio (the full version, _not_ Code). We are generally only interested in the functions under `postgres.exe`, as the library should load other DLLs as necessary.
```cmd
dumpbin /imports "C:/Program Files/PostgreSQL/15/lib/LIBRARY_NAME.dll"
```
## Linux
On Linux, we make use of the built-in `nm` command. We are interested in the `U` functions that do not have an `@` near the end (as those are usually implemented in external libraries).
```bash
nm -D -u /usr/lib/postgresql/15/lib/LIBRARY_NAME.so
```
