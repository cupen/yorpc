start:
	tmux new-session -d -s yorpc
	tmux setenv -t yorpc HISTFILE /dev/null
	tmux split-window -t "yorpc:0"   -v -p 50
	tmux split-window -t "yorpc:0.1" -h -p 50
	tmux send-keys    -t "yorpc:0.0" "go run main.go --server" Enter
	tmux send-keys    -t "yorpc:0.1" "sleep 1 && go run main.go --client" Enter
	tmux select-pane  -t "yorpc:0.1" 
	tmux attach       -t yorpc
	tmux kill-session -t yorpc


stop:
	tmux kill-session -t yorpc

