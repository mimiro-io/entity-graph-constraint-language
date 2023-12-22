package egcl

import (
	"fmt"
	egdm "github.com/mimiro-io/entity-graph-data-model"
	"github.com/pkg/errors"
	"math"
)

type DataProvider interface {
	Hop(sourceEntityId string, reference string, inverse bool) ([]*egdm.Entity, error)
	GetEntity(entityId string) (*egdm.Entity, error)
	GetDatasetEntities(name string) ([]*egdm.Entity, error)
}

type SchemaValidator interface {
	ValidateEntity(schema *Schema, entity *egdm.Entity) (ok bool, exceptions []*ConstraintViolation, err error)
	ValidateDataset(schema *Schema, datasetName string) (ok bool, exceptions []*ConstraintViolation, err error)
	ValidateSchema(schema *Schema) (ok bool, exceptions []*ConstraintViolation, err error)
	ValidateEntityCollection(schema *Schema, entityCollection *egdm.EntityCollection) (ok bool, exceptions []*ConstraintViolation, err error)
}

type NamespaceResolver interface {
	ResolveNamespace(prefix string) (string, error)
}

type ValidatorSettings struct {
	// StrictValidation will cause the validator to fail if an entity has properties or references that are not defined in the schema
	StrictValidation bool
	// ValidateRelated enables validation of referenced entities, specifically that they exist and are of the correct type
	ValidateRelated bool
	// datasetsContext defines the set of datasets that are in play when checking referenced entities
	DatasetsContext []string
}

type Validator struct {
	settings     *ValidatorSettings
	dataProvider DataProvider
}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) WithSettings(settings *ValidatorSettings) *Validator {
	v.settings = settings
	return v
}

func (v *Validator) WithDataProvider(provider DataProvider) *Validator {
	v.dataProvider = provider
	return v
}

type ViolationType int

const (
	MinPropertyOccurrenceNotMet ViolationType = iota
	MaxPropertyOccurrenceExceeded
	MinReferenceOccurrenceNotMet
	MaxReferenceOccurrenceExceeded
	ReferenceNotFound
	ReferenceTypeMismatch
	AbstractEntityClassViolation
)

// for our purposes, we can use math.MaxInt32 as infinity
const MaxInt = math.MaxInt32

type ConstraintViolation struct {
	Constraint    any
	Entity        *egdm.Entity
	ViolationType ViolationType
	Message       string
}

func NewConstraintViolation(constraint any, entity *egdm.Entity, violationType ViolationType, message string) *ConstraintViolation {
	violation := &ConstraintViolation{}
	violation.Constraint = constraint
	violation.Entity = entity
	violation.ViolationType = violationType
	violation.Message = message
	return violation
}

// ValidateDataset validates the given dataset against the data in the named dataset
func (v *Validator) ValidateDataset(schema *Schema, datasetName string) (ok bool, exceptions []*ConstraintViolation, err error) {
	//TODO implement me
	panic("implement me")
}

// ValidateSchema validates the given schema against data accessible to this validator
func (v *Validator) ValidateSchema(schema *Schema) (ok bool, exceptions []*ConstraintViolation, err error) {
	//TODO implement me

	// return v.CheckIsAbstractConstraint(c)

	panic("implement me")
}

// ValidateEntityCollection validates the given entity collection against the given schema
func (v *Validator) ValidateEntityCollection(schema *Schema, entityCollection *egdm.EntityCollection) (ok bool, exceptions []*ConstraintViolation, err error) {
	exceptions = make([]*ConstraintViolation, 0)
	ok = true
	err = nil

	for _, entity := range entityCollection.Entities {
		valid, entityExceptions, err := v.ValidateEntity(schema, entity)
		if err != nil {
			return false, nil, err
		}

		if !valid {
			ok = false
			exceptions = append(exceptions, entityExceptions...)
		}
	}

	return
}

func (v *Validator) ValidateEntity(schema *Schema, entity *egdm.Entity) (ok bool, exceptions []*ConstraintViolation, err error) {
	exceptions = make([]*ConstraintViolation, 0)
	ok = true
	err = nil

	// get entity classes
	classes, ok := entity.References[RDfTypeURI]
	if !ok {
		return
	}

	classArray := makeStringArray(classes)
	if classArray == nil {
		return
	}

	for _, class := range classArray {
		constraints := schema.GetConstraintsForEntityClass(class, true)
		for _, constraint := range constraints {
			// check if constraint is violated
			valid, violation, constraintError := v.CheckConstraint(schema, constraint, entity)
			if constraintError != nil {
				err = constraintError
				return
			}

			if !valid && violation != nil {
				ok = false
				exceptions = append(exceptions, violation)
			}
		}
	}

	return
}

// CheckConstraint checks if the given constraint is violated for the given entity. Returns false if the constraint is ok
// and true and a constraint violation struct if not. Error is returned if something went wrong while checking the constraint.
func (v *Validator) CheckConstraint(schema *Schema, constraint any, entity *egdm.Entity) (bool, *ConstraintViolation, error) {
	switch c := constraint.(type) {
	case *ReferenceConstraint:
		return v.CheckReferenceConstraint(entity, c)
	case *InverseReferenceConstraint:
		return v.CheckInverseReferenceConstraint(entity, c)
	case *PropertyConstraint:
		return v.CheckPropertyConstraint(entity, c)
	case *IsAbstractConstraint:
		// this checks if this entity is an implementation of a EntityClass that is defined as abstract
		return v.CheckEntityIsAbstractConstraint(schema, entity, c)
	}

	return false, nil, errors.New("constraint type not supported")
}

func (v *Validator) CheckReferenceConstraint(entity *egdm.Entity, constraint *ReferenceConstraint) (bool, *ConstraintViolation, error) {
	propertyURI, err := constraint.GetConstrainedPropertyClass()
	if err != nil {
		return false, nil, err
	}
	minCard := constraint.GetMinAllowedOccurrences()
	maxCard := constraint.GetMaxAllowedOccurrences()
	if maxCard == -1 {
		maxCard = MaxInt
	}

	value, ok := entity.References[propertyURI]
	if ok {
		// check cardinality
		switch values := value.(type) {
		case []string:
			if len(values) < minCard {
				cv := NewConstraintViolation(constraint, entity, MinReferenceOccurrenceNotMet,
					fmt.Sprintf("min card is %d but found %d occurrences", minCard, len(values)))
				return false, cv, nil
			} else if len(values) > maxCard {
				cv := NewConstraintViolation(constraint, entity, MaxReferenceOccurrenceExceeded,
					fmt.Sprintf("max card is %d but found %d occurrences", maxCard, len(values)))
				return false, cv, nil
			}

			// if enabled check related entity exists and is of correct type
			if v.settings.ValidateRelated {
				if v.dataProvider == nil {
					return false, nil, errors.New("no data provider configured")
				}

				for _, ref := range values {
					allowedReferencedClass, err := constraint.GetAllowedReferencedClass()
					if err != nil {
						return false, nil, err
					}
					valid, cv, err := v.CheckExistenceAndTypeOfReferencedEntity(ref, allowedReferencedClass)
					if err != nil {
						return false, nil, err
					}
					if !valid {
						// this isnt ideal but slightly better than passing these in just to get added to the violation
						cv.Entity = entity
						cv.Constraint = constraint
						return false, cv, nil
					}
				}
			}

		default:
			if minCard > 1 {
				cv := NewConstraintViolation(constraint, entity, MinReferenceOccurrenceNotMet,
					fmt.Sprintf("min card is %d but found %d occurrences", minCard, 1))
				return false, cv, nil
			}
		}
	} else {
		// check that the property is optional
		if minCard > 0 {
			v := NewConstraintViolation(constraint, entity, MinReferenceOccurrenceNotMet,
				fmt.Sprintf("min card is %d but found %d occurrences", minCard, 0))
			return false, v, nil
		}
	}

	return true, nil, nil
}

func (v *Validator) CheckExistenceAndTypeOfReferencedEntity(entityId string, expectedType string) (valid bool, violation *ConstraintViolation, err error) {
	entity, err := v.dataProvider.GetEntity(entityId)
	if err != nil {
		return false, nil, err
	}

	if entity == nil {
		return false, NewConstraintViolation(nil, nil, ReferenceNotFound, fmt.Sprintf("related entity with id %v", entityId)), nil
	}

	// entity is partials so need to look at each one and check the dataset it belongs to and then if correct check the type
	partials := entity.Properties["http://data.mimiro.io/core/partials"]
	if partials == nil {
		return false, nil, errors.New("no partials found")
	}

	// iterate over partials and check if the dataset is correct
	if partialsArray, ok := partials.([]*egdm.Entity); ok {
		for _, partial := range partialsArray {
			entityDataset := partial.Properties["http://data.mimiro.io/core/dataset"].(string)
			if v.isDatasetInContext(entityDataset) {
				// check type
				entityType := partial.Properties["http://www.w3.org/1999/02/22-rdf-syntax-ns#type"].(string)
				if entityType == expectedType {
					return true, nil, nil
				} else {
					return false, NewConstraintViolation(nil, nil, ReferenceTypeMismatch, fmt.Sprintf("expected type %v but found %v", expectedType, entityType)), nil
				}
			}
		}
	}

	return false, NewConstraintViolation(nil, nil, ReferenceNotFound, fmt.Sprintf("related entity with id %v", entityId)), nil
}

func (v *Validator) isDatasetInContext(dataset string) bool {
	if v.settings.DatasetsContext == nil {
		return true
	} else {
		for _, ds := range v.settings.DatasetsContext {
			if ds == dataset {
				return true
			}
		}
		return false
	}
}

func (v *Validator) CheckInverseReferenceConstraint(entity *egdm.Entity, constraint *InverseReferenceConstraint) (bool, *ConstraintViolation, error) {
	return true, nil, nil
}

func (v *Validator) CheckPropertyConstraint(entity *egdm.Entity, constraint *PropertyConstraint) (valid bool, violation *ConstraintViolation, err error) {
	propertyURI, err := constraint.GetConstrainedPropertyClass()
	if err != nil {
		return false, nil, err
	}
	minCard := constraint.GetMinAllowedOccurrences()
	maxCard := constraint.GetMaxAllowedOccurrences()
	if maxCard == -1 {
		maxCard = MaxInt
	}

	value, ok := entity.Properties[propertyURI]
	if ok {
		// check cardinality
		switch values := value.(type) {
		case []any:
			if len(values) < minCard {
				cv := NewConstraintViolation(constraint, entity, MinPropertyOccurrenceNotMet,
					fmt.Sprintf("min card is %d but found %d occurrences", minCard, len(values)))
				return false, cv, nil
			} else if len(values) > maxCard {
				cv := NewConstraintViolation(constraint, entity, MaxPropertyOccurrenceExceeded,
					fmt.Sprintf("max card is %d but found %d occurrences", maxCard, len(values)))
				return false, cv, nil
			}
		default:
			if minCard > 1 {
				cv := NewConstraintViolation(constraint, entity, MinPropertyOccurrenceNotMet,
					fmt.Sprintf("min card is %d but found %d occurrences", minCard, 1))
				return false, cv, nil
			}
		}

		// todo: check value data type and also value pattern if specified
	} else {
		// check that the property is optional
		if minCard > 0 {
			v := NewConstraintViolation(constraint, entity, MinPropertyOccurrenceNotMet,
				fmt.Sprintf("min card is %d but found %d occurrences", minCard, 0))
			return false, v, nil
		}
	}

	return true, nil, nil
}

func (v *Validator) CheckEntityIsAbstractConstraint(schema *Schema, entity *egdm.Entity, constraint *IsAbstractConstraint) (bool, *ConstraintViolation, error) {
	// get the rdf type of the entity provided
	classes, ok := entity.References[RDfTypeURI]
	if !ok {
		// no rdf type found, so this entity is not instance if an abstract class
		return true, nil, nil
	}

	// get class in the constraint
	abstractClass := constraint.GetEntityClass()

	classesArray := makeStringArray(classes)
	if classesArray != nil {
		for _, class := range classesArray {
			if class == abstractClass {
				return false, NewConstraintViolation(constraint, entity, AbstractEntityClassViolation,
					fmt.Sprintf("expected type %v to be abstract but entity %s is an instance", class, entity.ID)), nil
			}
		}
	}

	return true, nil, nil
}

func (v *Validator) CheckGlobalIsAbstractConstraint(constraint *IsAbstractConstraint) (bool, *ConstraintViolation, error) {
	if v.dataProvider == nil {
		return false, nil, errors.New("no reference resolver configured")
	}

	// get all instances of specified type
	abstractType, err := constraint.Entity.GetFirstReferenceValue(EGCLentityClass)
	if err != nil {
		return false, nil, err
	}

	instances, err := v.dataProvider.Hop(abstractType, RDfTypeURI, true)
	if err != nil {
		return false, nil, err
	}

	if len(instances) > 0 {
		return false, NewConstraintViolation(constraint, constraint.Entity, ReferenceTypeMismatch,
			fmt.Sprintf("expected type %v to have 0 instances but found %d", abstractType, len(instances))), nil
	}

	return true, nil, nil
}

func makeStringArray(val interface{}) []string {
	switch v := val.(type) {
	case []string:
		return v
	case string:
		res := make([]string, 1)
		res[0] = v
		return res
	default:
		return nil
	}
}
