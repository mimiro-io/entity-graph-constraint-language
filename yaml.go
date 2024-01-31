package egcl

import (
	"fmt"
	egdm "github.com/mimiro-io/entity-graph-data-model"
	"gopkg.in/yaml.v3"
)

func parseYaml(data []byte) (*Schema, error) {
	yamlSchema := make([]map[string]any, 0)
	err := yaml.Unmarshal(data, &yamlSchema)
	if err != nil {
		return nil, err
	}

	// entity collection to contain all constraint entities
	nsm := egdm.NewNamespaceContext()
	ec := egdm.NewEntityCollection(nsm)

	yamlNamespaces, ok := yamlSchema[0]["namespaces"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("error parsing context")
	}

	for key, value := range yamlNamespaces {
		nsm.StorePrefixExpansionMapping(key, value.(string))
	}

	constraintCount := 0

	yamlClasses := yamlSchema[1:]
	for _, classData := range yamlClasses {
		entityClass := egdm.NewEntity()

		// add rdf type of egcl class
		entityClass.SetReference("rdf:type", "egcl:EntityClass")

		// get id property  from classData
		id, ok := classData["id"]
		if !ok {
			return nil, fmt.Errorf("error class id is missing")
		}
		entityClass.SetID(id.(string))

		for key, value := range classData {
			switch key {
			case "isAbstract":
				abstractConstraint := egdm.NewEntity()
				abstractConstraint.SetID(fmt.Sprintf("%s-constraint-%d", entityClass.ID, constraintCount))
				abstractConstraint.SetReference("rdf:type", "egcl:IsAbstractConstraint")
				abstractConstraint.SetReference("egcl:entityClass", entityClass.ID)
				// add to constraints
				err := ec.AddEntity(abstractConstraint)
				if err != nil {
					return nil, err
				}
				constraintCount++
			case "superclasses":
				superClasses := make([]string, 0)
				for _, v := range value.([]interface{}) {
					superClasses = append(superClasses, v.(string))
				}
				entityClass.SetReference("egcl:subClassOf", superClasses)
			case "label":
				entityClass.SetProperty("egcl:label", value.(string))
			case "description":
				entityClass.SetProperty("egcl:description", value.(string))
			case "refs":
				for k, v := range value.(map[interface{}]interface{}) {
					entityClass.SetReference(k.(string), v.(string))
				}
			case "props":
				for k, v := range value.(map[interface{}]interface{}) {
					entityClass.SetProperty(k.(string), v.(string))
				}
			case "propertyConstraints":
				for _, v := range value.([]any) {
					propConstraint := egdm.NewEntity()
					propConstraint.SetID(fmt.Sprintf("%s-constraint-%d", entityClass.ID, constraintCount))
					propConstraint.SetReference("rdf:type", "egcl:PropertyConstraint")
					propConstraint.SetReference("egcl:entityClass", entityClass.ID)
					constraintCount++
					for k, val := range v.(map[string]any) {
						switch k {
						case "propertyClass":
							propConstraint.SetReference("egcl:propertyClass", val.(string))
						case "datatype":
							propConstraint.SetProperty("egcl:datatype", val.(string))
						case "minCard":
							propConstraint.SetProperty("egcl:minCard", val.(int))
						case "maxCard":
							propConstraint.SetProperty("egcl:maxCard", val.(int))
						}
					}

					err := ec.AddEntity(propConstraint)
					if err != nil {
						return nil, err
					}
				}
			case "referenceConstraints":
				for _, v := range value.([]any) {
					refConstraint := egdm.NewEntity()
					refConstraint.SetID(fmt.Sprintf("%s-constraint-%d", entityClass.ID, constraintCount))
					refConstraint.SetReference("rdf:type", "egcl:ReferenceConstraint")
					refConstraint.SetReference("egcl:entityClass", entityClass.ID)
					constraintCount++
					for k, val := range v.(map[string]any) {
						switch k {
						case "referenceClass":
							refConstraint.SetReference("egcl:referenceClass", val.(string))
						case "referencedEntityClass":
							refConstraint.SetReference("egcl:referencedEntityClass", val.(string))
						case "minCard":
							refConstraint.SetProperty("egcl:minCard", val.(int))
						case "maxCard":
							refConstraint.SetProperty("egcl:maxCard", val.(int))
						case "inverseReferenceClass":
							refConstraint.SetReference("egcl:inverseReferenceClass", val.(string))
						case "inverseMinCard":
							refConstraint.SetProperty("egcl:inverseMinCard", val.(int))
						case "inverseMaxCard":
							refConstraint.SetProperty("egcl:inverseMaxCard", val.(int))
						}
					}
					err := ec.AddEntity(refConstraint)
					if err != nil {
						return nil, err
					}
				}
			}
		}

		// add class entity to collection
		err := ec.AddEntity(entityClass)
		if err != nil {
			return nil, err
		}
	}

	// expand prefixes
	err = ec.ExpandNamespacePrefixes()
	if err != nil {
		return nil, err
	}

	return NewSchema(ec), nil
}
