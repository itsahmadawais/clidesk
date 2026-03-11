# CLIDesk bash/zsh wrapper
# Launches clidesk and cd's to whatever directory you ended up in.
#
# SETUP — add this to your ~/.bashrc or ~/.zshrc:
#
#   source /path/to/clidesk/shell/clidesk.sh
#
# Then just type:  clidesk
# Or with a theme: clidesk --theme nord

clidesk() {
    local tmp
    tmp=$(mktemp)
    clidesk.bin "$@" --print-dir "$tmp"
    local dir
    dir=$(cat "$tmp" 2>/dev/null)
    rm -f "$tmp"
    if [ -n "$dir" ] && [ -d "$dir" ]; then
        cd "$dir" || return
    fi
}
