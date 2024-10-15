# lol

location of log

This is a very simple, but practical library for logging in applications. Its
main feature is printing source code locations to make debugging easier.

## terminals

Due to how so few terminals actually support source location hyperlinks, pretty much tilix and intellij terminal are 
the only two that really provide adequate functionality, this logging library defaults to output format that works 
best with intellij. As such, the terminal is aware of the CWD and the code locations printed are relative, as 
required to get the hyperlinkization from this terminal. Handling support for Tilix requires more complications and 
due to advances with IntelliJ's handling it is not practical to support any other for this purpose. Users of this 
library can always fall back to manually interpreting and accessing the relative file path to find the source of a log.

In addition, due to this terminal's slow rendering of long lines, long log strings are automatically broken into 80 
character lines, and if there is comma separators in the line, the line is broken at the comma instead of at column80. This works perfectly for this purpose.