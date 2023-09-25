# clinote

TUI notes manager to read, edit, and add notes in style

### Install

```
go install github.com/L-Michael1/clinote@latest
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

To configure the path to where your notes are stored - change the $NOTES_FOLDER path

```
export NOTES_FOLDER="CUSTOM_PATH"
```

To configure your editor:

```
export EDITOR="CUSTOM_EDITOR"
```
