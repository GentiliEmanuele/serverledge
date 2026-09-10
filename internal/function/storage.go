package function

type Storage interface {
	Get(name string) (*Function, bool)
	Save(function *Function) error
	Delete(f *Function) error
	GetAllWithPrefix(prefix string) ([]string, error)
	GetAll() ([]string, error)
}

var storage Storage

func InitStorage(cfg string) {
	switch cfg {
	case "garage":
		storage = &GarageStorage{}
	case "etcd":
		storage = &EtcdStorage{}
	default:
		storage = nil
	}
}

func GetStorage() Storage {
	// Return garage storage as default
	if storage == nil {
		storage = &GarageStorage{}
	}

	return storage
}
