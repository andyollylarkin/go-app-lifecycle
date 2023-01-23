package application

import (
	"context"
	"github.com/rs/zerolog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// TODO: get service
type Application struct {
	services map[string]*Service // user defined services
	//TODO вынести в отдельный тип service locator
	sysSignals  []syscall.Signal
	mainFunc    MainFunc
	recoverFunc RecoverFunc
	appState    ApplicationState
	shutdownCh  chan struct{}
	// time for gracefully application shutting down
	shutdownTimeout time.Duration
	waitFunc        func()
	// After this exceeded will force app terminate.
	initTimeout time.Duration
	logger      zerolog.Logger //TODO: direct dependency from zerolog logger. Extract logger interface
}

// CreateApplication create new application
func CreateApplication(logger zerolog.Logger, shutdownTimeout time.Duration, initTimeout time.Duration,
	mainFunc MainFunc, recoverFunc RecoverFunc) Application {
	return Application{appState: StateStart, logger: logger, shutdownTimeout: shutdownTimeout,
		initTimeout: initTimeout, mainFunc: mainFunc, recoverFunc: recoverFunc, services: make(map[string]*Service)}
}

// Run application lifecycle. Must be called from main function
func (app *Application) Run() error {
	quitChan := app.initSysSignals()
	//shutdownCh := make(chan struct{})
	app.shutdownCh = make(chan struct{})
	app.waitFunc = app.wait()
	timeoutCtx, cancelInit := context.WithTimeout(context.Background(), app.initTimeout)
	defer cancelInit()
	appStage := NewApplicationStage()

	keeper := NewServiceKeeper(app.services, 0, 0)
	app.logger.Info().Msg("Start initialization phase.")
	err := appStage.Init(timeoutCtx, &app.appState, keeper)
	if err != nil {
		app.logger.Error().Msgf("Error while initialization: %s", err.Error())
		return err
	}
	if err == nil {
		app.logger.Info().Msg("The initialization phase was successful.")
	}

	healthCheckErrCh := make(chan error)
	if err == nil {
		go func() {
			err = keeper.HealthCheck()
			healthCheckErrCh <- err
		}()
	}

	appErrCh := make(chan error)
	if err == nil {
		app.logger.Info().Msg("Try starting app...")
		go func() {
			defer func() {
				if r := recover(); r != nil {
					if m, ok := r.(string); ok {
						app.logger.Error().Msgf("Main application panic: %s", m)
					}
					if app.recoverFunc != nil {
						err = app.recoverFunc()
						if err != nil {
							app.logger.Error().Msgf("Error while recover: %s", err.Error())
						}
						close(appErrCh)
					}
				}
			}()
			err = appStage.Start(context.TODO(), &app.appState, keeper, app.mainFunc, app.waitFunc)
			appErrCh <- err

		}()
		app.logger.Info().Msg("App started.")
	}

	select {
	case err = <-appErrCh:
		if err != nil {
			app.logger.Error().Msgf("Main application error: %s", err.Error())
		}
		return app.shutdownHandler(appStage)
	case err = <-healthCheckErrCh:
		if err != nil {
			app.logger.Error().Msgf("Health check error: %s", err.Error())
		}
		return app.shutdownHandler(appStage)
	case sig := <-quitChan:
		app.logger.Info().Msgf("Received signal: %s. Waiting for graceful completion.", sig.String())
		return app.shutdownHandler(appStage)
	}
}

func (app *Application) wait() func() {
	return func() {
		<-app.shutdownCh
	}
}

// SetSysSignals append os signal(s) to application
// TODO: check if windows. Signals works on windows?
func (app *Application) SetSysSignals(signals ...syscall.Signal) *Application {
	for _, s := range signals {
		app.sysSignals = append(app.sysSignals, s)
	}
	return app
}

func (app *Application) RegisterService(serviceName string, service Service) *Application {
	app.services[serviceName] = &service
	return app
}

func (app *Application) initSysSignals() (sysQuitSignal chan os.Signal) {
	// if the user has not set which signals he wants to process, set default signals
	if len(app.sysSignals) == 0 {
		app.sysSignals = []syscall.Signal{syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGINT}
	}
	sysQuitSignal = make(chan os.Signal)
	for _, s := range app.sysSignals {
		signal.Notify(sysQuitSignal, s)
	}
	return sysQuitSignal
}

func (app *Application) gracefulShutdownApp(appStage *Stage) error {
	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), app.shutdownTimeout)
	defer cancelShutdown()
	app.logger.Info().Msg("Wait graceful shutdown.")
	err := appStage.Shutdown(ctxShutdown, &app.appState, app.shutdownCh)
	if err != nil {
		app.logger.Err(err)
		return err
	}
	app.logger.Info().Msg("Graceful shutdown completed.")
	app.logger.Info().Msg("Resource uninitialization.")
	uninitTimeoutCtx, uninitCancel := context.WithTimeout(context.Background(), app.initTimeout)
	defer uninitCancel()
	err = appStage.Uninit(uninitTimeoutCtx, &app.appState)
	if err != nil {
		app.logger.Error().Msgf("Error when uninitialization: %s", err)
		return err
	}
	app.logger.Info().Msg("Resource uninitialization completed.")
	app.logger.Info().Msg("Application has completed")
	return err
}

func (app *Application) shutdownHandler(appStage *Stage) error {
	shutdownErr := app.gracefulShutdownApp(appStage)
	if shutdownErr != nil {
		app.logger.Error().Msgf("Error while shutdown: %s", shutdownErr.Error())
	}
	return shutdownErr
}
