package models

import (
	"github.com/ottemo/foundation/env"
)

// GetModelAndSetID retrieves current model implementation and sets its ID to some value
func GetModelAndSetID(modelName string, modelID string) (InterfaceStorable, error) {
	someModel, err := GetModel(modelName)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	storableModel, ok := someModel.(InterfaceStorable)
	if !ok {
		return nil, env.ErrorNew(ConstErrorModule, ConstErrorLevel, "652d1949b661438e9097231a52734feb", "model is not InterfaceStorable capable")
	}

	err = storableModel.SetID(modelID)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	return storableModel, nil
}

// LoadModelByID loads model data in current implementation
func LoadModelByID(modelName string, modelID string) (InterfaceStorable, error) {

	someModel, err := GetModel(modelName)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	storableModel, ok := someModel.(InterfaceStorable)
	if !ok {
		return nil, env.ErrorNew(ConstErrorModule, ConstErrorLevel, "56ec204adbb949fca5e99d43e1f19025", "model is not InterfaceStorable capable")
	}

	err = storableModel.Load(modelID)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	return storableModel, nil
}
