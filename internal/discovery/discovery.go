
package discovery

import (
	"context"
	"fmt"
	"time"

	"github.com/beijian128/pineapple/internal/utils"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

type ServiceInfo struct {
	Name    string
	Addr    string
	Port    int
	Version string
}

type Discovery struct {
	client *clientv3.Client
	config *utils.EtcdConfig
	lease  clientv3.LeaseID
}

var GlobalDiscovery *Discovery

func NewDiscovery(cfg *utils.EtcdConfig) (*Discovery, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   cfg.Endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client: %w", err)
	}

	return &Discovery{
		client: client,
		config: cfg,
	}, nil
}

func InitDiscovery(cfg *utils.EtcdConfig) error {
	var err error
	GlobalDiscovery, err = NewDiscovery(cfg)
	if err != nil {
		return err
	}

	utils.Logger.Info("etcd connected successfully",
		zap.Strings("endpoints", cfg.Endpoints))

	return nil
}

func CloseDiscovery() {
	if GlobalDiscovery != nil && GlobalDiscovery.client != nil {
		_ = GlobalDiscovery.client.Close()
		utils.Logger.Info("etcd disconnected")
	}
}

func (d *Discovery) Register(service *ServiceInfo) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	leaseResp, err := d.client.Grant(ctx, d.config.LeaseTTL)
	if err != nil {
		return fmt.Errorf("failed to grant lease: %w", err)
	}
	d.lease = leaseResp.ID

	key := fmt.Sprintf("/services/%s/%s:%d", service.Name, service.Addr, service.Port)
	value := fmt.Sprintf(`{"name":"%s","addr":"%s","port":%d,"version":"%s"}`,
		service.Name, service.Addr, service.Port, service.Version)

	_, err = d.client.Put(ctx, key, value, clientv3.WithLease(leaseResp.ID))
	if err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}

	keepAliveChan, err := d.client.KeepAlive(context.Background(), leaseResp.ID)
	if err != nil {
		return fmt.Errorf("failed to keep alive: %w", err)
	}

	go func() {
		for ka := range keepAliveChan {
			_ = ka
		}
	}()

	utils.Logger.Info("service registered",
		zap.String("name", service.Name),
		zap.String("addr", service.Addr),
		zap.Int("port", service.Port))

	return nil
}

func (d *Discovery) Unregister(service *ServiceInfo) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("/services/%s/%s:%d", service.Name, service.Addr, service.Port)
	_, err := d.client.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to unregister service: %w", err)
	}

	if d.lease != 0 {
		_, _ = d.client.Revoke(ctx, d.lease)
	}

	utils.Logger.Info("service unregistered",
		zap.String("name", service.Name))

	return nil
}

func (d *Discovery) Discover(serviceName string) ([]*ServiceInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	prefix := fmt.Sprintf("/services/%s/", serviceName)
	resp, err := d.client.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to discover services: %w", err)
	}

	var services []*ServiceInfo
	for _, _ = range resp.Kvs {
		svc := &ServiceInfo{}
		services = append(services, svc)
	}

	return services, nil
}
