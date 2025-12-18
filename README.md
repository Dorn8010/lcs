# lcs
Library Command Search tool for CLI commands</br>
Store and find long commands easily</br>
============================================
A CLI based simple, fast and sleek command library tool, to store complex commands with descriptions and shortcuts.</br>
It can also ask for user defined variables.</br>
</br>
© 2025 by Alexander Dorn - MIT license</br>
</br>
## Usage: lcs [option] search_term_always_last</br>

============================================================</br>
Searches for a command in the description and offers the findings for execution</br>
</br>
The DB contains an explanation and the command with optional variables ~/.lcs-db.csv is a ; separated CSV</br>
Here an example of an entry - which is generated automatically when you use the --add option :</br>
Echo test;echo "Hello World"</br>
</br>
### Options</br>
  --help, -h     Show this help</br>
  --version      Show version info</br>
  --verbose, -v  Show verbose logging</br>
  --fast, -f     Fast select option number (e.g. -f 2)</br>
  --print        Print command only</br>
  --copy         Copy command to clipboar (no execution)</br>
  --add          Add a new command. Usage: lcs --add "Desc" "Cmd" or interactive mode</br>
  --edit         Search, remove and re-add/edit</br>
  --remove       Search and remove a command</br>
  --db           Path to custom database, default: ~/.lcs-db.csv</br>
</br>
### Using Variables</br>
  You can define variables in commands to be filled at runtime.</br>
  Syntax: {"Label":"DefaultValue"}</br>
  Example: ssh -i {"KeyFile":"~/.ssh/id_rsa"} user@host</br>
</br>
<b>Examples</b></br>
Add a command :           lcs --add "Echo test" 'echo "Hello World"'</br>
Execute/recall :          lcs hello</br>
Copy cmd to clipboard:    lcs --copy hello</br>
Remove command:           lcs --remove hello</br>
Print/View cmd:           lcs --print hello</br>
</br>
## Installation in Mac or Linux (e.g. Ubuntu)</br>

### Testing
For testing no installation is needed, just make the bash version executable with</br>
chmod +x lcs.sh</br>
If you like the concept, compile the fast and compact GO version with full functionality.

### To compile on Linux</br>
  sudo apt install golang && go build -o lcs main.go</br>
  ./lcs --help</br>
To install locally</br>
  sudo cp lcs /usr/local/bin/</br>

### To compile on Mac</br>
  brew install go && go build -o lcs main.go</br>
  ./lcs --help</br>
To install locally</br>
  sudo cp lcs /usr/local/bin/</br>
