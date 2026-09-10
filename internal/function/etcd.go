package function

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/serverledge-faas/serverledge/utils"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdStorage struct{}

func (es *EtcdStorage) Get(name string) (*Function, bool) {
	cli, err := utils.GetEtcdClient()
	if err != nil {
		return nil, false
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	getResponse, err := cli.Get(ctx, getEtcdKey(name))
	if err != nil {
		utils.TriggerEtcdReconnection()
		log.Printf("etcd get failed: %v", err)
		return nil, false
	} else if len(getResponse.Kvs) < 1 {
		return nil, false
	}

	var f Function
	err = json.Unmarshal(getResponse.Kvs[0].Value, &f)
	if err != nil {
		return nil, false
	}

	return &f, true
}

func (es *EtcdStorage) Save(f *Function) error {
	cli, err := utils.GetEtcdClient()
	if err != nil {
		return err
	}
	ctx := context.TODO()

	payload, err := json.Marshal(*f)
	if err != nil {
		return fmt.Errorf("could not marshal function: %v", err)
	}
	_, err = cli.Put(ctx, f.getEtcdKey(), string(payload))
	if err != nil {
		utils.TriggerEtcdReconnection()
		return fmt.Errorf("failed Put: %v", err)
	}

	return nil
}

func (es *EtcdStorage) Delete(f *Function) error {
	cli, err := utils.GetEtcdClient()
	if err != nil {
		return err
	}
	ctx := context.TODO()

	deleteResponse, err := cli.Delete(ctx, f.getEtcdKey())
	if err != nil {
		return fmt.Errorf("failed Delete: %v", err)
	} else if deleteResponse.Deleted != 1 {
		fmt.Printf("no function with key '%s' exists", f.getEtcdKey())
	}

	return nil
}

func (es *EtcdStorage) GetAll() ([]string, error) {
	return GetStorage().GetAllWithPrefix("/function")
}

func (es *EtcdStorage) GetAllWithPrefix(prefix string) ([]string, error) {
	cli, err := utils.GetEtcdClient()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()

	resp, err := cli.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	functions := make([]string, len(resp.Kvs))
	for i, s := range resp.Kvs {
		functions[i] = string(s.Key)[len(prefix+"/"):]
	}

	return functions, ctx.Err()
}
