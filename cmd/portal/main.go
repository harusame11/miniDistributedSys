package main

import (
	"My_mimiDistributed/log"
	"My_mimiDistributed/portal"
	"My_mimiDistributed/registry"
	"My_mimiDistributed/service"
	"context"
	"fmt"
	stlog "log"
)

func main() {
	err := portal.ImportTemplates()
	if err != nil {
		stlog.Fatal(err)
	}
	host, port := "localhost", "5000"
	serviceAddress := fmt.Sprintf("http://%s:%s", host, port)
	r := registry.Registration{
		ServiceName: registry.PortalService,
		ServiceURL:  serviceAddress,
		RequireServices: []registry.ServiceName{
			registry.LogService,
			registry.GradingService,
		},
		ServiceUpdateURL: serviceAddress + "/services",
	}

	ctx, err := service.Start(context.Background(),
		r,
		host,
		port,
		portal.RegisterHandlers)
	if err != nil {
		stlog.Fatal(err)
	}
	//为客户端设定logger
	if logProvider, err := registry.GetProvider(registry.LogService); err == nil {
		log.SetClientLogger(logProvider, r.ServiceName)
	}
	<-ctx.Done()
	fmt.Println("Shutting down portal")
}
