package hw06pipelineexecution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	currentChannel := in
	for _, stage := range stages {
		currentChannel = RunStage(currentChannel, done, stage)
	}
	return currentChannel
}

func RunStage(in In, done In, stage Stage) Out {
	out := make(Bi)
	go func() {
		defer close(out)
		stageOut := stage(in)
		for {
			select {
			case _, ok := <-done:
				if !ok {
					// дренируем каналы чтобы быстрее завершилась работа
					for v := range in {
						_ = v
					}
					for v := range stageOut {
						_ = v
					}
					return
				}
			case v, ok := <-stageOut:
				if !ok {
					return
				}
				out <- v
			}
		}
	}()
	return out
}
