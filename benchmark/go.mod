module benchmark

go 1.25.5

replace github.com/msrexe/patron => ../

require (
	github.com/gammazero/workerpool v1.1.3
	github.com/msrexe/patron v0.0.0-00010101000000-000000000000
	github.com/panjf2000/ants/v2 v2.11.3
	github.com/sourcegraph/conc v0.3.0
	golang.org/x/sync v0.19.0
)

require (
	github.com/gammazero/deque v0.2.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.9.0 // indirect
)
