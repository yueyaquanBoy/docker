// +build windows

package winresources

/*

This package is for embedding a manifest file and an icon into docker.exe.
The benefit of this is that a manifest file does not need to be alongside
the .exe, and there is an icon when docker runs. The real plus to this is
that the tool to embed the resources is cross-platform.

The file containng the resources is winresources\rsrc.syso. To regenerate it,
use rsrc.exe from https://github.com/akavel/rsrc

Commands:
go get github.com/akavel/rsrc
Cd [pathto]src\github.com\akavel\rsrc
Go build
Copy rsrc.exe to somewhere on your path
From docker\hack\make run makewindowsresources.cmd
Move the resulting rsrc.syso file to the winresources directory


The file will be automatically picked up by go build, no post-processing
steps are required.

*/
