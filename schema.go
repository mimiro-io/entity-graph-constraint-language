package egcl

import (
	egdm "github.com/mimiro-io/entity-graph-data-model"
	errors "github.com/pkg/errors"
)

const RDFUriExpansion = "http://www.w3.org/1999/02/22-rdf-syntax-ns#"
const EGCLUriExpansion = "http://data.mimiro.io/egcl/"

const RDfTypeURI = RDFUriExpansion + "type"

const EGCLEntityClass = EGCLUriExpansion + "EntityClass"

const EGCLPropertyConstraint = EGCLUriExpansion + "PropertyConstraint"
const EGCLReferenceConstraint = EGCLUriExpansion + "ReferenceConstraint"
const EGCLApplicationConstraint = EGCLUriExpansion + "ApplicationConstraint"
const EGCLIsAbstractConstraint = EGCLUriExpansion + "IsAbstractConstraint"

const EGCLentityClass = EGCLUriExpansion + "entityClass"
const EGCLpropertyClass = EGCLUriExpansion + "propertyClass"
const EGCLreferenceClass = EGCLUriExpansion + "referenceClass"
const EGCLInverseReferenceClass = EGCLUriExpansion + "inverseReferenceClass"
const EGCLallowedReferencedClass = EGCLUriExpansion + "referencedEntityClass"
const EGCLsubclassOf = EGCLUriExpansion + "subClassOf"

const EGCLdatatype = EGCLUriExpansion + "datatype"

const EGCLLabel = EGCLUriExpansion + "label"
const EGCLDescription = EGCLUriExpansion + "description"

const EGCLAny = EGCLUriExpansion + "Any"

const (
	EGCLsortable              = EGCLUriExpansion + "sortable"
	EGCLqueryable             = EGCLUriExpansion + "queryable"
	EGCLisunique              = EGCLUriExpansion + "isunique"
	EGCLminCardinality        = EGCLUriExpansion + "minCard"
	EGCLmaxCardinality        = EGCLUriExpansion + "maxCard"
	EGCLrule                  = EGCLUriExpansion + "rule"
	EGCLinverseMinCardinality = EGCLUriExpansion + "inverseMinCard"
	EGCLinverseMaxCardinality = EGCLUriExpansion + "inverseMaxCard"
)

var (
	ErrMissingParent = errors.New("referenced super class is not defined in the schema")
)

func NewSchema(entities *egdm.EntityCollection) *Schema {
	schema := &Schema{
		EntityCollection: entities,
		EntityClasses:    make([]*EntityClass, 0),
		Constraints:      make([]any, 0),
	}
	schema.initialize()
	return schema
}

type Schema struct {
	EntityCollection *egdm.EntityCollection
	EntityClasses    []*EntityClass
	Constraints      []any
	baseURI          string
	description      string
}

func (aSchema *Schema) IsOfType(entity *egdm.Entity, typeURI string) bool {
	if entity == nil {
		return false
	}

	types := entity.References[RDfTypeURI]
	switch t := types.(type) {
	case []string:
		for _, tt := range t {
			entityTypeURI := tt
			if entityTypeURI == typeURI {
				return true
			}
		}
	case string:
		if t == typeURI {
			return true
		}
	}
	return false
}

func (aSchema *Schema) GetEntityClassById(entityClassIdentifier string) *EntityClass {
	for _, ec := range aSchema.EntityClasses {
		if ec.Entity.ID == entityClassIdentifier {
			return ec
		}
		if uri, err := aSchema.EntityCollection.NamespaceManager.GetFullURI(ec.Entity.ID); err == nil { // expand it
			if uri == entityClassIdentifier {
				return ec
			}
		}

	}
	return nil
}

func (aSchema *Schema) GetEntityClassClassHierarchy(entityClassIdentifier string) ([]string, error) {
	ancestors := make([]string, 0)
	entityClass := aSchema.GetEntityClassById(entityClassIdentifier)
	if entityClass == nil {
		return nil, errors.Wrap(ErrMissingParent, entityClassIdentifier)
	}
	parentClass, _ := entityClass.Entity.GetFirstReferenceValue(EGCLsubclassOf)
	var err2 error
	for parentClass != "" {
		ancestors = append(ancestors, parentClass)
		parentEntityClass := aSchema.GetEntityClassById(parentClass)
		if parentEntityClass == nil {
			err2 = errors.Wrap(ErrMissingParent, parentClass)
			break
		}
		parentClass, _ = parentEntityClass.Entity.GetFirstReferenceValue(EGCLsubclassOf)
	}
	if err2 != nil {
		return nil, err2
	}
	return ancestors, nil
}

// Get any reference constraints that are inverse and thus outgoing for the specific entityClassIdentifer
func (aSchema *Schema) GetOutgoingInverseConstraintsForEntityClass(entityClassIdentifier string, inherited bool) []any {
	constraints := make([]any, 0)
	for _, candidate := range aSchema.Constraints {
		switch constraint := candidate.(type) {
		case *ReferenceConstraint:

			appliesToEntityClass, err := constraint.Entity.GetFirstReferenceValue(EGCLallowedReferencedClass)
			if err == nil {
				_, err := constraint.GetInverseConstrainedPropertyClass()
				if err == nil {
					if appliesToEntityClass == entityClassIdentifier {
						constraints = append(constraints, constraint)
					}
				}
			}
		}
	}

	if inherited {
		superclasses, _ := aSchema.GetEntityClassClassHierarchy(entityClassIdentifier)
		for _, superClass := range superclasses {
			superConstraints := aSchema.GetOutgoingInverseConstraintsForEntityClass(superClass, false)
			constraints = append(constraints, superConstraints...)
		}
	}

	return constraints
}

func (aSchema *Schema) GetConstraintsForEntityClass(entityClassIdentifier string, inherited bool) []any {
	constraints := make([]any, 0)
	for _, candidate := range aSchema.Constraints {
		switch constraint := candidate.(type) {
		case *ReferenceConstraint:
			appliesToEntityClass, err := constraint.Entity.GetFirstReferenceValue(EGCLentityClass)
			if err == nil {
				if appliesToEntityClass == entityClassIdentifier {
					constraints = append(constraints, constraint)
				}
			}
		case *PropertyConstraint:
			appliesToEntityClass, err := constraint.Entity.GetFirstReferenceValue(EGCLentityClass)
			if err == nil {
				if appliesToEntityClass == entityClassIdentifier {
					constraints = append(constraints, constraint)
				}
			}
		case *IsAbstractConstraint:
			appliesToEntityClass, err := constraint.Entity.GetFirstReferenceValue(EGCLentityClass)
			if err == nil && appliesToEntityClass != "" {
				if appliesToEntityClass == entityClassIdentifier {
					constraints = append(constraints, constraint)
				}
			}
		case *ApplicationConstraint:
			appliesToEntityClass, err := constraint.Entity.GetFirstReferenceValue(EGCLentityClass)
			if err == nil && appliesToEntityClass != "" {
				if appliesToEntityClass == entityClassIdentifier {
					constraints = append(constraints, constraint)
				}
			}
		}

	}

	// go up the hierarchy and get other constraints
	if inherited {
		superclasses, _ := aSchema.GetEntityClassClassHierarchy(entityClassIdentifier)
		for _, superClass := range superclasses {
			superConstraints := aSchema.GetConstraintsForEntityClass(superClass, false)
			constraints = append(constraints, superConstraints...)
		}
	}

	return constraints
}

func (aSchema *Schema) IsAbstract(entityClass *EntityClass) bool {
	constraints := aSchema.GetConstraintsForEntityClass(entityClass.Entity.ID, false)
	for _, candidate := range constraints {
		switch candidate.(type) {
		case *IsAbstractConstraint:
			return true
		}
	}
	return false
}

func (aSchema *Schema) initialize() {
	entities := aSchema.EntityCollection.Entities
	for _, entity := range entities {
		if aSchema.IsOfType(entity, EGCLEntityClass) {
			ec := newEntityClass(entity)
			aSchema.EntityClasses = append(aSchema.EntityClasses, ec)
		} else if aSchema.IsOfType(entity, EGCLPropertyConstraint) {
			c := newPropertyConstraint(entity)
			aSchema.Constraints = append(aSchema.Constraints, c)
		} else if aSchema.IsOfType(entity, EGCLReferenceConstraint) {
			c := newReferenceConstraint(entity)
			aSchema.Constraints = append(aSchema.Constraints, c)
		} else if aSchema.IsOfType(entity, EGCLIsAbstractConstraint) {
			c := newIsAbstractConstraint(entity)
			aSchema.Constraints = append(aSchema.Constraints, c)
		} else if aSchema.IsOfType(entity, EGCLApplicationConstraint) {
			c := newApplicationConstraint(entity)
			aSchema.Constraints = append(aSchema.Constraints, c)
		}
	}
}

func newEntityClass(entity *egdm.Entity) *EntityClass {
	ec := &EntityClass{}
	ec.Entity = entity
	return ec
}

type EntityClass struct {
	Entity *egdm.Entity
}

func (entityClass *EntityClass) GetLabel() string {
	val, err := entityClass.Entity.GetFirstStringPropertyValue(EGCLLabel)
	if err != nil {
		return ""
	}
	return val
}

func (entityClass *EntityClass) GetDescription() string {
	val, err := entityClass.Entity.GetFirstStringPropertyValue(EGCLDescription)
	if err != nil {
		return ""
	}
	return val
}

type Constraint struct {
	Entity                   *egdm.Entity
	ConstraintTypeIdentifier string
}

type QueryConstraint struct {
	Constraint
}

func newPropertyConstraint(entity *egdm.Entity) *PropertyConstraint {
	pc := &PropertyConstraint{}
	pc.Constraint.Entity = entity
	pc.Constraint.ConstraintTypeIdentifier = EGCLPropertyConstraint
	return pc
}

type PropertyConstraint struct {
	Constraint
}

func (c *PropertyConstraint) GetDataType() string {
	val, err := c.Entity.GetFirstReferenceValue(EGCLdatatype)
	if err != nil {
		return EGCLAny
	}
	return val
}

func (c *PropertyConstraint) GetMinAllowedOccurrences() int {
	val, err := c.Entity.GetFirstIntPropertyValue(EGCLminCardinality)
	if err != nil {
		return 0
	}
	return val
}

func (c *PropertyConstraint) GetMaxAllowedOccurrences() int {
	val, err := c.Entity.GetFirstIntPropertyValue(EGCLmaxCardinality)
	if err != nil {
		return -1
	}
	return val
}

func (c *PropertyConstraint) GetConstrainedPropertyClass() (string, error) {
	return c.Entity.GetFirstReferenceValue(EGCLpropertyClass)
}

func (c *PropertyConstraint) GetConstrainedEntityClass() (string, error) {
	return c.Entity.GetFirstReferenceValue(EGCLentityClass)
}

func (c *PropertyConstraint) GetSortable() string {
	if res, err := c.Entity.GetFirstStringPropertyValue(EGCLsortable); err == nil {
		return res
	}
	return ""
}

func (c *PropertyConstraint) GetIsQueryable() bool {
	if res, err := c.Entity.GetFirstBooleanPropertyValue(EGCLqueryable); err == nil {
		return res
	}
	return false
}

func (c *PropertyConstraint) GetIsUnique() bool {
	if res, err := c.Entity.GetFirstBooleanPropertyValue(EGCLisunique); err == nil {
		return res
	}
	return false
}

func newReferenceConstraint(entity *egdm.Entity) *ReferenceConstraint {
	rc := &ReferenceConstraint{}
	rc.Constraint.Entity = entity
	rc.Constraint.ConstraintTypeIdentifier = EGCLReferenceConstraint
	return rc
}

type ReferenceConstraint struct {
	Constraint
}

func (aReferenceConstraint *ReferenceConstraint) GetMinAllowedOccurrences() int {
	val, err := aReferenceConstraint.Entity.GetFirstIntPropertyValue(EGCLminCardinality)
	if err != nil {
		return 0
	}
	return val
}

func (aReferenceConstraint *ReferenceConstraint) GetConstrainedEntityClass() (string, error) {
	return aReferenceConstraint.Entity.GetFirstReferenceValue(EGCLentityClass)
}

func (aReferenceConstraint *ReferenceConstraint) GetMaxAllowedOccurrences() int {
	val, err := aReferenceConstraint.Entity.GetFirstIntPropertyValue(EGCLmaxCardinality)
	if err != nil {
		return -1
	}
	return val
}

func (aReferenceConstraint *ReferenceConstraint) GetConstrainedPropertyClass() (string, error) {
	return aReferenceConstraint.Entity.GetFirstReferenceValue(EGCLreferenceClass)
}

func (aReferenceConstraint *ReferenceConstraint) GetAllowedReferencedClass() (string, error) {
	return aReferenceConstraint.Entity.GetFirstReferenceValue(EGCLallowedReferencedClass)
}

func (aReferenceConstraint *ReferenceConstraint) GetInverseConstrainedPropertyClass() (string, error) {
	return aReferenceConstraint.Entity.GetFirstReferenceValue(EGCLInverseReferenceClass)
}

func newIsAbstractConstraint(entity *egdm.Entity) *IsAbstractConstraint {
	iac := &IsAbstractConstraint{}
	iac.Constraint.Entity = entity
	iac.Constraint.ConstraintTypeIdentifier = EGCLIsAbstractConstraint
	return iac
}

type IsAbstractConstraint struct {
	Constraint
}

type ApplicationConstraint struct {
	Constraint
}

func (c *ApplicationConstraint) GetRule() string {
	if res, err := c.Entity.GetFirstReferenceValue(EGCLrule); err == nil {
		return res
	}
	return ""
}

func (c *ApplicationConstraint) Match(namespaceManager egdm.NamespaceManager, rule string) bool {
	m := c.GetRule()
	if rule == m {
		return true
	}

	// probably have a curie, so expand it to full url and check against that
	curie, _ := namespaceManager.GetFullURI(rule)
	return curie == m
}

func newApplicationConstraint(entity *egdm.Entity) *ApplicationConstraint {
	c := &ApplicationConstraint{}
	c.Constraint.Entity = entity
	c.Constraint.ConstraintTypeIdentifier = EGCLApplicationConstraint

	return c
}

type PropertyValueConstraint struct {
}

type ConstraintViolation struct {
}

type PropertyConstraintViolation struct {
}
