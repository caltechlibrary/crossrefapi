
Installation
============

*crossrefapi* is a command line program run from a shell like Bash. It allows you to retrieve JSON works and types object from the CrossRefAPI in a
[polite](https://api.crossref.org/swagger-ui/index.html "look under the heading Etiquette") manner.

Quick install with curl
-----------------------

If you are running macOS or Linux you can install released versions of crossrefapi
with the following curl command.

~~~
curl https://caltechlibrary.github.io/crossrefapi/installer.sh
~~~

On Windows the install command in Powershell is 

~~~
irm https://caltechlibrary.github.io/crossrefapi/installer.ps1 | iex
~~~

Compiled version
----------------

Compiled versions are available for macOS (Intel and M1 processor, macOS-x86_64, macOS-arm64), Linux (Intel process, Linux-x86_64), 
Windows (Intel processor, Windows-x86_64) and Rapsberry Pi (arm7 processor, RaspberryPiOS-arm7)

VERSION_NUMBER is a [symantic version number](http://semver.org/) (e.g. v0.1.2)


For all the released version go to the project page on Github and click latest release

>    https://github.com/caltechlibrary/crossrefapi/releases/latest


| Platform      | Zip Filename                             |
|---------------|------------------------------------------|
| Windows       | crossrefapi-VERSION_NUMBER-Windows-x86_64.zip |
| macOS (Intel) | crossrefapi-VERSION_NUMBER-macOS-x86_64.zip  |
| macOS (M1,M2) | crossrefapi-VERSION_NUMBER-macOS-arm64.zip  |
| Linux (Intel) | crossrefapi-VERSION_NUMBER-Linux-x86_64.zip   |
| Linux (ARM64) | crossrefapi-VERSION_NUMBER-Linux-aarch64.zip   |
| Raspbery Pi OS (64) | crossrefapi-VERSION_NUMBER-RaspberryPiOS-arm7.zip |


The basic recipe
----------------

+ Find the Zip file listed matching the architecture you're running and download it
    + (e.g. if you're on a Windows 10 laptop/Surface with a Intel style CPU you'd choose the Zip file with "Windows-x86_64" in the name).
+ Download the zip file and unzip the file.
+ Copy the contents of the folder named "bin" to a folder that is in your path 
    + (e.g. "$HOME/bin" is common).
+ Adjust your PATH if needed
    + (e.g. export PATH="$HOME/bin:$PATH")
+ Test


### macOS

1. Download the zip file
2. Unzip the zip file
3. Copy the executables to $HOME/bin (or a folder in your path)
4. Make sure the new location in in our path
5. Test

Here's an example of the commands run in the Terminal App after downloading the 
zip file.

```shell
    cd Downloads/
    unzip crossrefapi-*-macOS-x86_64.zip
    mkdir -p $HOME/bin
    cp -v bin/* $HOME/bin/
    export PATH=$HOME/bin:$PATH
    crossrefapi -version
```

### Windows

1. Download the zip file
2. Unzip the zip file
3. Copy the executables to $HOME/bin (or a folder in your path)
4. Test

Here's an example of the commands run in from the Bash shell on Windows 10 after
downloading the zip file.

```shell
    cd Downloads/
    unzip crossrefapi-*-Windows-x86_64.zip
    mkdir -p $HOME/bin
    cp -v bin/* $HOME/bin/
    export PATH=$HOME/bin:$PATH
    crossrefapi -version
```


### Linux 

1. Download the zip file
2. Unzip the zip file
3. Copy the executables to $HOME/bin (or a folder in your path)
4. Test

Here's an example of the commands run in from the Bash shell after
downloading the zip file.

```shell
    cd Downloads/
    unzip crossrefapi-*-Linux-x86_64.zip
    mkdir -p $HOME/bin
    cp -v bin/* $HOME/bin/
    export PATH=$HOME/bin:$PATH
    crossrefapi -version
```


### Raspberry Pi

Released version is for a Raspberry Pi 2 or later use (i.e. requires ARM 7 support).

1. Download the zip file
2. Unzip the zip file
3. Copy the executables to $HOME/bin (or a folder in your path)
4. Test

Here's an example of the commands run in from the Bash shell after
downloading the zip file.

```shell
    cd Downloads/
    unzip crossrefapi-*-RaspberryPiOS-arm7.zip
    mkdir -p $HOME/bin
    cp -v bin/* $HOME/bin/
    export PATH=$HOME/bin:$PATH
    crossrefapi -version
```


Compiling from source
---------------------

_crossrefapi_ is "go gettable".  Use the "go get" command to download the dependant packages
as well as _crossrefapi_'s source code. 


```shell
    go get -u github.com/caltechlibrary/crossrefapi/...
```

Or clone the repstory and then compile

```shell
    cd
    git clone https://github.com/caltechlibrary/crossrefapi src/github.com/caltechlibrary/crossrefapi
    cd src/github.com/caltechlibrary/crossrefapi
    make
    make test
    make install
```

Compilation assumes [go](https://github.com/golang/go) v1.16

