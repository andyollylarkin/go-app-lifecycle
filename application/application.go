package application

import (
	"context"
	"github.com/rs/zerolog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Application struct {
	services map[string]*Service // user defined services
	//TODO вынести в отдельный тип service locator
	sysSignals  []syscall.Signal
	mainFunc    MainFunc
	recoverFunc RecoverFunc
	appState    ApplicationState
	// time for gracefully application shutting down
	shutdownTimeout time.Duration
	// After this exceeded will force app terminate.
	initTimeout time.Duration
	logger      zerolog.Logger //TODO: direct dependency from zerolog logger. Extract logger interface
}

// CreateApplication create new application
func CreateApplication(logger zerolog.Logger, shutdownTimeout time.Duration, initTimeout time.Duration,
	mainFunc MainFunc, recoverFunc RecoverFunc) Application {
	return Application{appState: StateStart, logger: logger, shutdownTimeout: shutdownTimeout,
		initTimeout: initTimeout, mainFunc: mainFunc, recoverFunc: recoverFunc}
}

// Run application lifecycle. Must be called from main function
func (app *Application) Run() error {
	quitChan := app.initSysSignals()
	appTermination := make(chan struct{})
	shutdownCh := make(chan struct{})
	timeoutCtx, cancelInit := context.WithTimeout(context.Background(), app.initTimeout)
	defer cancelInit()
	appStage := NewApplicationStage()

	keeper := NewServiceKeeper(app.services, 0, 0)
	app.logger.Info().Msg("Start initialization phase.")
	err := appStage.Init(timeoutCtx, &app.appState, keeper)
	if err != nil {
		app.logger.Error().Msgf("Error while initialization: %s", err.Error())
		close(appTermination)
	}
	if err == nil {
		app.logger.Info().Msg("The initialization phase was successful.")
	}

	healthCheckCh := make(chan error)
	if err == nil {
		go func() {
			err = keeper.HealthCheck()
			healthCheckCh <- err
		}()
	}

	appErrCh := make(chan error)
	if err == nil {
		app.logger.Info().Msg("Try starting app...")
		go func() {
			defer func() {
				if r := recover(); r != nil {
					//TODO: тут возможно стоит возобновлять работу mainFunc, либо закрывать appTermination
					if m, ok := r.(string); ok {
						app.logger.Error().Msg(m)
					}
					if app.recoverFunc != nil {
						err := app.recoverFunc()
						if err != nil {
							app.logger.Error().Msgf("Error while recover: %s", err.Error())
						}
					}
				}
			}()
			err = appStage.Start(context.TODO(), &app.appState, keeper, shutdownCh, app.mainFunc)
			appErrCh <- err

		}()
		app.logger.Info().Msg("App started.")
	}

	select {
	case err = <-appErrCh:
		println("App error")
		err = app.gracefulShutdownApp(appStage, shutdownCh)
		app.logger.Err(err)
		return err
	case err = <-healthCheckCh:
		app.logger.Error().Msgf("Health check error: %s", err.Error())
		err = app.gracefulShutdownApp(appStage, shutdownCh)
		if err != nil {
			app.logger.Error().Msgf("App shutdown error: %s", err.Error())
		}
		return err
	case sig := <-quitChan:
		app.logger.Info().Msgf("Received signal: %s. Waiting for graceful completion.", sig.String())
		err = app.gracefulShutdownApp(appStage, shutdownCh)
		if err != nil {
			app.logger.Info().Msgf("App shutdown error: %s", err.Error())
		}
		//time.Sleep(app.shutdownTimeout)
		//close(shutdownCh)
		app.logger.Info().Msgf("Interrupt by system signal, %s", sig.String())
	case <-appTermination:
		err = app.gracefulShutdownApp(appStage, shutdownCh)
		return err
	}
	return nil
}

// SetSysSignals append os signal(s) to application
// TODO: check if windows. Signals works on windows?
func (app *Application) SetSysSignals(signals ...syscall.Signal) *Application {
	for _, s := range signals {
		app.sysSignals = append(app.sysSignals, s)
	}
	return app
}

func (app *Application) RegisterService(serviceName string, service *Service) *Application {
	app.services[serviceName] = service
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

func (app *Application) gracefulShutdownApp(appStage *ApplicationStage, shutdownCh chan struct{}) error {
	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), app.shutdownTimeout)
	defer cancelShutdown()
	app.logger.Info().Msg("Wait graceful shutdown.")
	err := appStage.Shutdown(ctxShutdown, &app.appState, shutdownCh)
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
