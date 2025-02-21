package monitor

func (bm *BlockMonitor) Stop() {
	close(bm.stopChan)
	bm.wg.Wait()
}
