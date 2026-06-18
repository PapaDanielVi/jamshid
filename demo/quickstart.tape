# Quick start: add a profile, link it, check status, unlink.
# Run with: vhs demo/quickstart.tape
Output demo/quickstart.gif

Set Shell "bash"
Set Width 1200
Set Height 700
Set FontSize 18
Set TypingSpeed 60ms
Set Padding 20

# Isolated, self-cleaning sandbox with a sample .claude config to import.
Hide
Type "export JAMSHID_DEMO=$(mktemp -d) && export HOME=$JAMSHID_DEMO && mkdir -p $HOME/project/.claude && cd $HOME/project && echo '{}' > .claude/settings.local.json && export PS1='$ ' && clear"
Enter
Sleep 800ms
Show

Type "jamshid add personal"
Sleep 500ms
Enter
Sleep 1.5s
Type "y"
Enter
Sleep 2s

Type "jamshid list"
Sleep 500ms
Enter
Sleep 2s

# The profile now owns the settings, so clear the local copy and let jamshid manage it.
Hide
Type "rm -rf .claude"
Enter
Show

Type "jamshid link personal"
Sleep 500ms
Enter
Sleep 2s

Type "jamshid status"
Sleep 500ms
Enter
Sleep 2.5s

Type "jamshid unlink"
Sleep 500ms
Enter
Sleep 2s

# Tear down the sandbox.
Hide
Type "cd / && rm -rf $JAMSHID_DEMO"
Enter
Show
