# clinote

TUI notes manager to read, edit, and add notes in style

![clinote_demo](https://github.com/L-Michael1/clinote/assets/27537005/d5384977-81ee-42e1-b605-483a5a53dfff)

### Install

```
go install github.com/L-Michael1/clinote@latest
```

### Path configuration

Ensure that you have the go bin directory in your path:
```
export PATH="$HOME/go/bin:$PATH"
```

### Run

```
clinote
```

### How to use

- `↑/k`: Navigate up a line on the table / Navigate up a line in your note
- `↓/j`: Navigate down a line on the table / Navigate down a line in your note
- `b/backspace/esc`: Go back to the table view
- `enter`: View the selected note
- `n`: Create a new note
- `e`: Edit the selected note
- `?`: Toggle the help menu
- `q`: Quit the program

### Folder & editor configuration

Default configuration:

```
$NOTES_FOLDER = $HOME + "/notes/"
$EDITOR = "vim"
```

To configure the path to where your notes are stored - change the $NOTES_FOLDER path:

```
export NOTES_FOLDER="CUSTOM_PATH"
```

To configure your editor:

```
export EDITOR="CUSTOM_EDITOR"
```
