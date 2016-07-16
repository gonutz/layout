Windows Layout Daemon
=====================

This program lets you move the active window to a corner using the arrow keys. Just hold Left Control+Left Windows keys and then press Up-Left with the arrow keys to move the window to the upper-left hand corner, for example.

Pressing Escape while holding the control keys will stop the daemon.

The control keys can easily be changed in the code, see the top of the `main.go` file.

Use the `layout_deamon.bat` batch file to re-build the program and start it in the background. I have a copy of that file on my desktop and start it when I want the functionality but you could also add it as a Windows auto-start program.