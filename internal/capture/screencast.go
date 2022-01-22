package capture

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"demodesk/neko/internal/capture/gst"
	"demodesk/neko/internal/types"
)

// timeout between intervals, when screencast pipeline is checked
const screencastTimeout = 5 * time.Second

type ScreencastManagerCtx struct {
	logger zerolog.Logger
	mu     sync.Mutex
	wg     sync.WaitGroup

	pipeline    *gst.Pipeline
	pipelineStr string
	pipelineMu  sync.Mutex

	image      types.Sample
	tickerStop chan struct{}

	enabled bool
	started bool
	expired int32
}

func screencastNew(enabled bool, pipelineStr string) *ScreencastManagerCtx {
	logger := log.With().
		Str("module", "capture").
		Str("submodule", "screencast").
		Logger()

	manager := &ScreencastManagerCtx{
		logger:      logger,
		pipelineStr: pipelineStr,
		tickerStop:  make(chan struct{}),
		enabled:     enabled,
		started:     false,
	}

	manager.wg.Add(1)

	go func() {
		defer manager.wg.Done()

		ticker := time.NewTicker(screencastTimeout)
		defer ticker.Stop()

		for {
			select {
			case <-manager.tickerStop:
				return
			case <-ticker.C:
				if manager.Started() && !atomic.CompareAndSwapInt32(&manager.expired, 0, 1) {
					manager.stop()
				}
			}
		}
	}()

	return manager
}

func (manager *ScreencastManagerCtx) shutdown() {
	manager.logger.Info().Msgf("shutdown")

	manager.destroyPipeline()

	close(manager.tickerStop)
	manager.wg.Wait()
}

func (manager *ScreencastManagerCtx) Enabled() bool {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	return manager.enabled
}

func (manager *ScreencastManagerCtx) Started() bool {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	return manager.started
}

func (manager *ScreencastManagerCtx) Image() ([]byte, error) {
	atomic.StoreInt32(&manager.expired, 0)

	err := manager.start()
	if err != nil && !errors.Is(err, types.ErrCapturePipelineAlreadyExists) {
		return nil, err
	}

	if manager.image.Data == nil {
		return nil, errors.New("image data not found")
	}

	return manager.image.Data, nil
}

func (manager *ScreencastManagerCtx) start() error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	if !manager.enabled {
		return errors.New("screencast not enabled")
	}

	err := manager.createPipeline()
	if err != nil {
		return err
	}

	manager.started = true
	return nil
}

func (manager *ScreencastManagerCtx) stop() {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	manager.started = false
	manager.destroyPipeline()
}

func (manager *ScreencastManagerCtx) createPipeline() error {
	manager.pipelineMu.Lock()
	defer manager.pipelineMu.Unlock()

	if manager.pipeline != nil {
		return types.ErrCapturePipelineAlreadyExists
	}

	var err error

	manager.logger.Info().
		Str("str", manager.pipelineStr).
		Msgf("creating pipeline")

	manager.pipeline, err = gst.CreatePipeline(manager.pipelineStr)
	if err != nil {
		return err
	}

	manager.pipeline.AttachAppsink("appsink")
	manager.pipeline.Play()

	// get first image
	var ok bool
	select {
	case manager.image, ok = <-manager.pipeline.Sample:
		if !ok {
			return errors.New("unable to get first image")
		}
	case <-time.After(1 * time.Second):
		return errors.New("timeouted while waiting for first image")
	}

	manager.wg.Add(1)

	go func() {
		manager.logger.Debug().Msg("started receiving images")
		defer manager.wg.Done()

		for {
			image, ok := <-manager.pipeline.Sample
			if !ok {
				manager.logger.Debug().Msg("stopped receiving images")
				return
			}

			manager.image = image
		}
	}()

	return nil
}

func (manager *ScreencastManagerCtx) destroyPipeline() {
	manager.pipelineMu.Lock()
	defer manager.pipelineMu.Unlock()

	if manager.pipeline == nil {
		return
	}

	manager.pipeline.Destroy()
	manager.logger.Info().Msgf("destroying pipeline")
	manager.pipeline = nil
}
