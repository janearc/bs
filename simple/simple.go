package simple

import (
	"bs/config"
	"github.com/janearc/sux/sux"
	"github.com/sirupsen/logrus"
	"path/filepath"
)

type SimpleService struct {
	log       *logrus.Logger
	transport *sux.Sux
	cf        *config.Config
}

func NewSimpleService(root string) *SimpleService {
	svc := &SimpleService{}

	// logging facility
	l := logrus.New()
	l.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	svc.log = l

	// becky configs
	cf := filepath.Join(root, "etc/bs/config.yml")
	vf := filepath.Join(root, "etc/bs/version.yml")
	sf := filepath.Join(root, "etc/bs/secrets.yml")

	// sux configs
	scf := filepath.Join(root, "etc/sux/config.yml")
	ssf := filepath.Join(root, "etc/sux/secrets.yml")

	// load becky configs
	c, ce := config.LoadConfig(cf, vf, sf)
	if ce != nil {
		l.WithError(ce).Fatalf("Failed to load config: %v", ce)
	}

	svc.cf = c

	// stand up sux
	state := sux.NewSux(scf, vf, ssf)

	if state == nil {
		logrus.Fatal("Failed to instantiate Sux object")
	} else {
		logrus.Infof("SUX init successful")
	}

	svc.transport = state

	svc.log.Infof("Simple Becky Service instantiated")
	return svc
}

func (s *SimpleService) Chat(query string) string {
	s.log.Infof("Query: [%s]", query)
	return "return"
}
