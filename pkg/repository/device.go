package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Todorov99/sensorapi/pkg/entity"
	"github.com/Todorov99/sensorapi/pkg/global"
	"github.com/Todorov99/sensorapi/pkg/repository/query"
	"github.com/Todorov99/sensorapi/pkg/server/config"
)

type DeviceRepository interface {
	GetDeviceNameByID(ctx context.Context, id int) (string, error)
	GetDeviceIDByName(ctx context.Context, deviceName string) (int32, error)
	AddDeviceSensors(ctx context.Context, deviceID, sensorID int32) error
	IRepository
}

type deviceRepository struct {
	postgreClient *sql.DB
}

func NewDeviceRepository() DeviceRepository {
	return &deviceRepository{
		postgreClient: config.GetDatabaseCfg().GetPostgreClient(),
	}
}

func (d *deviceRepository) GetAll(ctx context.Context) (interface{}, error) {
	repositoryLogger.Debug("Getting all devicess...")
	devices := []*entity.Device{}

	err := executeSelectQuery(ctx, query.GetAllDevices, d.postgreClient, &devices)
	if err != nil {
		return nil, err
	}

	for _, device := range devices {
		sensors := []entity.Sensor{}
		err = executeSelectQuery(ctx, query.GetAllSensorsByDeviceID, d.postgreClient, &sensors, device.ID)
		if err != nil {
			return nil, err
		}

		device.Sensors = append(device.Sensors, sensors...)
	}
	repositoryLogger.Debug("Devices successfully retrieved")
	return devices, nil
}

func (d *deviceRepository) Add(ctx context.Context, model interface{}) error {
	repositoryLogger.Info("Adding device with all predifined sensors...")
	device := model.(entity.Device)

	err := executeModifyingQuery(ctx, query.InsertDevice, d.postgreClient, device.Name, device.Description)
	if err != nil {
		return err
	}

	repositoryLogger.Info("Device sucessfully added")
	return nil
}

func (d *deviceRepository) AddDeviceSensors(ctx context.Context, deviceID, sensorID int32) error {
	return executeModifyingQuery(ctx, query.InsertDeviceSensors, d.postgreClient, deviceID, sensorID)
}
func (d *deviceRepository) Update(ctx context.Context, model interface{}) error {
	device := model.(entity.Device)
	return executeModifyingQuery(ctx, query.UpdateDevice, d.postgreClient, device.Name, device.Description, device.ID)
}

func (d *deviceRepository) GetByID(ctx context.Context, id int) (interface{}, error) {
	repositoryLogger.Debugf("Getting device by ID: %d", id)
	device := &entity.Device{}

	err := executeSelectQuery(ctx, query.GetDeviceByID, d.postgreClient, device, id)
	if err != nil {
		return nil, err
	}

	sensors := []entity.Sensor{}
	err = executeSelectQuery(ctx, query.GetAllSensorsByDeviceID, d.postgreClient, &sensors, device.ID)
	if err != nil {
		return nil, err
	}

	device.Sensors = append(device.Sensors, sensors...)
	repositoryLogger.Debug("Devices successfully retrieved")

	return device, nil
}

func (d *deviceRepository) Delete(ctx context.Context, id int) error {
	repositoryLogger.Infof("Deleting device with id: %q", id)
	return executeModifyingQuery(ctx, query.DeleteDevice, d.postgreClient, id)
}

func (d *deviceRepository) GetDeviceIDByName(ctx context.Context, deviceName string) (int32, error) {
	repositoryLogger.Infof("Getting device ID by name: %q", deviceName)
	var id int32
	err := executeSelectQuery(ctx, query.GetDeviceIDByName, d.postgreClient, &id, deviceName)
	if err != nil {
		return 0, fmt.Errorf("failed getting device with name: %q: %w", deviceName, err)
	}

	if id == 0 {
		return 0, global.ErrorObjectNotFound
	}

	return id, nil
}

func (d *deviceRepository) GetDeviceNameByID(ctx context.Context, id int) (string, error) {
	var deviceName string
	err := executeSelectQuery(ctx, query.GetDeviceNameByID, d.postgreClient, &deviceName, id)
	if err != nil {
		return "", err
	}

	if deviceName == "" {
		return "", global.ErrorObjectNotFound
	}

	return deviceName, nil
}
