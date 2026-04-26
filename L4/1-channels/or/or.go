package or

func Or(channels ...<-chan any) <-chan any {
	N := len(channels)
	switch N {
	case 0:
		return nil
	case 1:
		return channels[0]
	case 2:
		done := make(chan any)
		go func() {
			defer close(done)
			select {
			case <-channels[0]:
			case <-channels[1]:
			}
		}()
		return done

	default:
		divide := N / 2
		return Or(Or(channels[:divide]...), Or(channels[divide:]...))

	}
}
