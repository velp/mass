package modules

type ModuleInterface interface {
	Run(goroutines int)
	Stop(wait bool)
}
