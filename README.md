# lcs
Library Command Search
A CLI based simple, fast and sleek command library tool, to store complex commands with descriptions and shortcuts. Can ask for user defined variables.

Library Command Search tool for CLI commands
Store and find long commands easily
© 2025 by Alexander Dorn            MIT lic.
                                Version 0.91
============================================

Usage: lcs [option] search_term

============================================
Searches for a command in the descr.
and offers the findings for exec.

The DB contains an explanation and
the command with optional variables

~/.lcs-db.csv is a ; separated CSV
Here an example of an entry :
Echo test;echo "Hello

Options :
  --help, -h     Show this help
  --version      Show version info
  --verbose, -v  Show verbose logging
  --fast, -f     Fast select option number (e.g. -f 2)
  --print        Print command only
  --copy         Copy command to clipboard
                 (no execution)
  --add          Add a new command
                 Usage: lcs --add "Desc" "Cmd"
                 or interactive mode
  --edit         Search, remove and re-add/edit
  --remove       Search and remove a command
  --db           Path to custom database
                 default: ~/.lcs-db.csv

Using Variables:
  You can define variables in commands to be filled at runtime.
  Syntax: {"Label":"DefaultValue"}
  Example: ssh -i {"KeyFile":"~/.ssh/id_rsa"} user@host
