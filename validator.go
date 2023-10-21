package egcl

import (
	egdm "github.com/mimiro-io/entity-graph-data-model"
)

type Validator struct {
}

type ConstraintViolation struct {
}

func (v *Validator) Validate(schema *Schema, entity *egdm.Entity) (ok bool, exceptions []*ConstraintViolation, err error) {
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
			if v.isViolated(constraint, entity) {
				ok = false
				exceptions = append(exceptions, &ConstraintViolation{})
			}
		}
	}

	return
}

func (v *Validator) isViolated(constraint any, entity *egdm.Entity) bool {
	switch c := constraint.(type) {
	case *ReferenceConstraint:
		v.CheckReferenceConstraint(entity, c)
	case *InverseReferenceConstraint:
		v.CheckInverseReferenceConstraint(entity, c)
	case *PropertyConstraint:
		v.CheckPropertyConstraint(entity, c)
	case *IsAbstractConstraint:
		v.CheckIsAbstractConstraint(entity, c)
	}

	return false
}

type ValidationContext interface {
	// return a given entity by id in the named datasets
	GetEntity(id string, datasets []string) *egdm.Entity

	// return the number of entities with a given id in specified datasets
	GetEntityCount(id string, datasets []string) int

	// Given an entity id and a predicate, returns a list of entities that reference it
	GetReferencingEntities(id string, reference string) []*egdm.Entity
}

func (v *Validator) CheckReferenceConstraint(entity *egdm.Entity, constraint *ReferenceConstraint) {

}

func (v *Validator) CheckInverseReferenceConstraint(entity *egdm.Entity, constraint *InverseReferenceConstraint) {

}

func (v *Validator) CheckPropertyConstraint(entity *egdm.Entity, constraint *PropertyConstraint) (*ConstraintViolation, error) {
	_, err := constraint.GetConstrainedPropertyClass()
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (v *Validator) CheckIsAbstractConstraint(entity *egdm.Entity, constraint *IsAbstractConstraint) {

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
