# Overview: help and version.
# Run with: vhs demo/overview.tape
Output demo/overview.gif

Set Shell "bash"
Set Width 1200
Set Height 700
Set FontSize 18
Set TypingSpeed 60ms
Set Padding 20

# Isolated, self-cleaning sandbox so nothing touches the real config.
Hide
Type "export JAMSHID_DEMO=$(mktemp -d) && export HOME=$JAMSHID_DEMO && mkdir -p $HOME/project && cd $HOME/project && export PS1='$ ' && clear"
Enter
Sleep 800ms
Show

Type "jamshid --help"
Sleep 500ms
Enter
Sleep 2.5s

Type "jamshid version"
Sleep 500ms
Enter
Sleep 1.5s

Type "jamshid list"
Sleep 500ms
Enter
Sleep 2s

# Tear down the sandbox.
Hide
Type "cd / && rm -rf $JAMSHID_DEMO"
Enter
Show
