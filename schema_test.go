package egcl

import (
	"strings"
	"testing"
	"time"

	"github.com/franela/goblin"
	egdm "github.com/mimiro-io/entity-graph-data-model"
)

func newParser() egdm.Parser {
	namespaceManager := egdm.NewNamespaceContext()
	parser := egdm.NewEntityParser(namespaceManager)
	return parser
}

func TestParseEntities(t *testing.T) {
	g := goblin.Goblin(t)

	config := `[
				{ "id" : "@context",
					"namespaces" : {
						"mimiro-people" : "http://data.mimiro.io/people/",
						"_" : "http://data.mimiro.io/core/"
					}
				}, {
					"id" : "mimiro-people:homer",
					"props" : { "name" : "Homer Simpson" },
					"refs" : { "friends" : [ "mimiro-people:marge" , "mimiro-people:bart"] }
				}, {
					"id" : "@continuation",
					"token" : "next-20"
				} ]`

	g.Describe("parse entities", func() {
		g.It("should parse without issues", func() {
			reader := strings.NewReader(config)
			parser := newParser()
			ec, err := parser.LoadEntityCollection(reader)

			g.Assert(err).IsNil()
			g.Assert(len(ec.Entities)).Equal(1, "wrong number of entities")

			g.Assert(ec.NamespaceManager).IsNotNil()
		})

		g.It("should expand prefix", func() {
			reader := strings.NewReader(config)
			parser := newParser()
			ec, err := parser.LoadEntityCollection(reader)

			g.Assert(err).IsNil()

			expansion, err := ec.NamespaceManager.GetNamespaceExpansionForPrefix("mimiro-people")
			g.Assert(err).IsNil()

			g.Assert(expansion).Equal("http://data.mimiro.io/people/")
		})

		g.It("should have the correct id", func() {
			reader := strings.NewReader(config)
			parser := newParser()
			ec, err := parser.LoadEntityCollection(reader)

			g.Assert(err).IsNil()

			entity := ec.Entities[0]
			g.Assert(entity.ID).Equal("http://data.mimiro.io/people/homer")

			expandedId, err := ec.NamespaceManager.GetFullURI(entity.ID)
			g.Assert(err).IsNil()

			g.Assert(expandedId).Equal("http://data.mimiro.io/people/homer")
		})

		g.It("should have continuation token", func() {
			reader := strings.NewReader(config)
			parser := newParser()
			ec, err := parser.LoadEntityCollection(reader)

			g.Assert(err).IsNil()

			g.Assert(ec.Continuation).IsNotNil()
			g.Assert(ec.Continuation.Token).Equal("next-20")
		})
	})
}

func TestParseClassesSchema(t *testing.T) {
	g := goblin.Goblin(t)

	config := `[
				{ "id" : "@context",
					"namespaces" : {
						"mimiro-schema" : "http://data.mimiro.io/schema/",
						"egcl" : "http://data.mimiro.io/egcl/",
						"rdf" : "http://www.w3.org/1999/02/22-rdf-syntax-ns#"
					}
				}, 
				{
					"id" : "mimiro-schema:Person",
					"refs" : { "rdf:type" : [ "egcl:EntityClass" ] }
				},
				{
					"id" : "mimiro-schema:Company",
					"refs" : { "rdf:type" : [ "egcl:EntityClass" ] }
				}]`

	g.Describe("parse schema classes", func() {
		g.It("should return classes", func() {

			reader := strings.NewReader(config)

			parser := newParser()
			ec, err := parser.LoadEntityCollection(reader)

			g.Assert(err).IsNil()

			schema := NewSchema(ec)
			classes := schema.EntityClasses

			g.Assert(len(classes)).Equal(2)
		})
	})

}

func TestParsePropertyConstraint(t *testing.T) {
	config := `[
				{ "id" : "@context",
					"namespaces" : {
						"mimiro-schema" : "http://data.mimiro.io/schema/",
						"egcl" : "http://data.mimiro.io/egcl/",
						"rdf" : "http://www.w3.org/1999/02/22-rdf-syntax-ns#",
						"xsd" : "http://www.w3.org/2001/XMLSchema/"
					}
				}, 
				{
					"id" : "mimiro-schema:Person",
					"refs" : { "rdf:type" : [ "egcl:EntityClass" ] }
				},
				{
					"id" : "mimiro-schema:Company",
					"refs" : { "rdf:type" : [ "egcl:EntityClass" ] }
				},
				{
					"id" : "mimiro-schema:constraint-1",
					"refs" : { 
							"rdf:type" : [ "egcl:PropertyConstraint" ],
							"egcl:entityClass" : "mimiro-schema:Person",
							"egcl:propertyClass" : "mimiro-schema:dateOfBirth",
							"egcl:datatype" : "xsd:DateTime"
					},
					"props" : {
							"egcl:minCard" : 0,
							"egcl:maxCard" : -1
					}
				}]`

	g := goblin.Goblin(t)
	g.Describe("parse property constraint", func() {
		g.It("should correctly parse constraints", func() {

			reader := strings.NewReader(config)

			parser := newParser()
			ec, err := parser.LoadEntityCollection(reader)

			g.Assert(err).IsNil()

			schema := NewSchema(ec)
			constraints := schema.Constraints

			g.Assert(len(constraints)).Equal(1)

			c1 := constraints[0]
			switch constraint := c1.(type) {
			case *Constraint:
				g.Fatalf("expecting property constraint and got constraint")
			case *PropertyConstraint:
				g.Assert(constraint.ConstraintTypeIdentifier).Equal(EGCLPropertyConstraint)
				g.Assert(constraint.Entity).IsNotNil()

				datatype := constraint.GetDataType()
				g.Assert(datatype).Equal("http://www.w3.org/2001/XMLSchema/DateTime", "wrong datatype")

				minCard := constraint.GetMinAllowedOccurrences()
				g.Assert(minCard).Equal(0, "wrong minCard")

				maxCard := constraint.GetMaxAllowedOccurrences()
				g.Assert(maxCard).Equal(-1, "wrong maxCard")
			default:
				g.Fatalf("expected constraint of type PropertyConstraint")
			}
		})
	})

}

func TestIsAbstractConstraint(t *testing.T) {
	reader := strings.NewReader(`[
				{ "id" : "@context",
					"namespaces" : {
						"mimiro-schema" : "http://data.mimiro.io/schema/",
						"egcl" : "http://data.mimiro.io/egcl/",
						"rdf" : "http://www.w3.org/1999/02/22-rdf-syntax-ns#"
					}
				}, 
				{
					"id" : "mimiro-schema:Person",
					"refs" : { "rdf:type" : [ "egcl:EntityClass" ] }
				},
				{
					"id" : "mimiro-schema:Company",
					"refs" : { "rdf:type" : [ "egcl:EntityClass" ] }
				},
				{
					"id" : "mimiro-schema:constraint-1",
					"refs" : { 
							"rdf:type" : [ "egcl:IsAbstractConstraint" ],
							"egcl:entityClass" : "mimiro-schema:Person"
					}
				}]`)

	parser := newParser()
	ec, err := parser.LoadEntityCollection(reader)

	if err != nil {
		t.Fatalf("error parsing schema: " + err.Error())
	}

	schema := NewSchema(ec)
	constraints := schema.Constraints

	if len(constraints) != 1 {
		t.Fatalf("expected 1 constraint")
	}

	// get entity class by id
	class := schema.GetEntityClassById("http://data.mimiro.io/schema/Person")
	if class == nil {
		t.Fatalf("no class found")
	}

	isAbstract := schema.IsAbstract(class)
	if !isAbstract {
		t.Fatalf("expected isabstract to be true")
	}
}

func TestEntityClassInheritance(t *testing.T) {
	config := `[
				{ "id" : "@context",
					"namespaces" : {
						"mimiro-schema" : "http://data.mimiro.io/schema/",
						"egcl" : "http://data.mimiro.io/egcl/",
						"rdf" : "http://www.w3.org/1999/02/22-rdf-syntax-ns#"
					}
				}, 

				{
					"id" : "mimiro-schema:Thing",
					"refs" : { "rdf:type" : [ "egcl:EntityClass" ] }
				},
				{
					"id" : "mimiro-schema:OrgUnit",
					"refs" : { 
								"rdf:type" : [ "egcl:EntityClass" ],
								"egcl:subClassOf" : "mimiro-schema:Thing"
							 }
				},
				{
					"id" : "mimiro-schema:Person",
					"refs" : { 
								"rdf:type" : [ "egcl:EntityClass" ],
								"egcl:subClassOf" : "mimiro-schema:OrgUnit"
							 }
				},
				{
					"id" : "mimiro-schema:Company",
					"refs" : { "rdf:type" : [ "egcl:EntityClass" ], 
								"egcl:subClassOf" : "mimiro-schema:OrgUnit"
 					}
				},
				{
					"id" : "mimiro-schema:constraint-1",
					"refs" : { 
							"rdf:type" : [ "egcl:PropertyConstraint" ],
							"egcl:entityClass" : "mimiro-schema:OrgUnit",
							"egcl:propertyClass" : "mimiro-schema:displayName"	
					},
					"props" : {
							"egcl:minCard" : 0,
							"egcl:maxCard" : 1
					}
				}]`

	g := goblin.Goblin(t)
	g.Describe("entity class inheritance", func() {
		g.It("should have a super class", func() {
			reader := strings.NewReader(config)

			parser := newParser()
			ec, err := parser.LoadEntityCollection(reader)

			g.Assert(err).IsNil()

			schema := NewSchema(ec)
			constraints := schema.Constraints

			g.Assert(len(constraints)).Equal(1)

			// get class hierarchy for person (expected orgunit)
			supers, err := schema.GetEntityClassClassHierarchy("http://data.mimiro.io/schema/Person")
			g.Assert(err).IsNil()
			g.Assert(len(supers)).Equal(2)
		})
		g.It("should have an org unit constraint", func() {
			reader := strings.NewReader(config)

			parser := newParser()
			ec, err := parser.LoadEntityCollection(reader)

			g.Assert(err).IsNil()

			schema := NewSchema(ec)
			constraints := schema.Constraints

			g.Assert(len(constraints)).Equal(1)

			// get class hierarchy for person (expected orgunit)
			supers, err := schema.GetEntityClassClassHierarchy("http://data.mimiro.io/schema/Person")
			g.Assert(err).IsNil()

			orgunitId := supers[0]
			g.Assert(orgunitId).Equal("http://data.mimiro.io/schema/OrgUnit")

			orgunitConstraints := schema.GetConstraintsForEntityClass("http://data.mimiro.io/schema/OrgUnit", true)
			g.Assert(len(orgunitConstraints)).Equal(1)

			thingId := supers[1]

			g.Assert(thingId).Equal("http://data.mimiro.io/schema/Thing")
		})
		g.It("should have a person constraint", func() {
			reader := strings.NewReader(config)

			parser := newParser()
			ec, err := parser.LoadEntityCollection(reader)

			g.Assert(err).IsNil()

			schema := NewSchema(ec)
			constraints := schema.Constraints

			g.Assert(len(constraints)).Equal(1)

			personConstraints := schema.GetConstraintsForEntityClass("http://data.mimiro.io/schema/Person", true)
			g.Assert(len(personConstraints)).Equal(1)
		})
	})

}

func TestInverseConstraint(t *testing.T) {
	config := `[
		{ "id" : "@context",
			"namespaces" : {
				"cima" : "http://data.mimiro.io/cima/",
				"egcl" : "http://data.mimiro.io/egcl/",
				"rdf" : "http://www.w3.org/1999/02/22-rdf-syntax-ns#"
			}
		}, 
		{
			"id": "cima:Farm",
			"refs": {
				"rdf:type": ["egcl:EntityClass"]
			},
			"props": {
				"egcl:label": "Farm",
				"egcl:description": "The logical grouping of things at a location/ account"
			}
		},
		{
        "id": "cima:User",
			"refs": {
				"rdf:type": ["egcl:EntityClass"]
			},
			"props": {
				"egcl:label": "User",
				"egcl:description": "A user of the farm system"
			}
		},
		{
			"id": "cima:User-constraint-1",
			"refs": {
				"rdf:type": "egcl:ReferenceConstraint",
				"egcl:entityClass": "cima:User",
				"egcl:referencedEntityClass": "cima:Farm",
				"egcl:referenceClass": "cima:operatesOn",
				"egcl:inverseReferenceClass": "cima:operator"
			},
			"props": {
				"egcl:minCard": 1,
				"egcl:maxCard": 1,
				"egcl:inverseMinCard": 1,
				"egcl:inverseMaxCard": -1
			}
		}]`

	g := goblin.Goblin(t)
	g.Describe("make sure inverse works", func() {
		g.It("should have the correct type", func() {
			reader := strings.NewReader(config)

			parser := newParser()
			ec, err := parser.LoadEntityCollection(reader)
			g.Assert(err).IsNil()
			schema := NewSchema(ec)

			entityClass := schema.GetEntityClassById("http://data.mimiro.io/cima/Farm")

			g.Assert(schema.IsOfType(entityClass.Entity, "http://data.mimiro.io/cima/Farm"))
		})
		g.It("should have 1 constraint", func() {
			reader := strings.NewReader(config)

			parser := newParser()
			ec, err := parser.LoadEntityCollection(reader)
			g.Assert(err).IsNil()

			schema := NewSchema(ec)
			constraints := schema.Constraints

			g.Assert(len(constraints)).Equal(1)

		})
		g.It("should have inverse property defined", func() {
			g.Timeout(time.Second * 600)
			reader := strings.NewReader(config)

			parser := newParser()
			ec, err := parser.LoadEntityCollection(reader)
			g.Assert(err).IsNil()

			schema := NewSchema(ec)
			userConstraints := schema.GetConstraintsForEntityClass("http://data.mimiro.io/cima/User", true)
			g.Assert(len(userConstraints)).Equal(1)

			c := userConstraints[0].(*ReferenceConstraint)
			invProp, _ := c.GetInverseConstrainedPropertyClass()
			g.Assert(invProp).Equal("http://data.mimiro.io/cima/operator")
		})
		g.It("inverse ref should appear as outgoing", func() {
			g.Timeout(time.Second * 600)
			reader := strings.NewReader(config)

			parser := newParser()
			ec, err := parser.LoadEntityCollection(reader)
			g.Assert(err).IsNil()

			schema := NewSchema(ec)
			userConstraints := schema.GetOutgoingInverseConstraintsForEntityClass("http://data.mimiro.io/cima/Farm", true)
			g.Assert(len(userConstraints)).Equal(1)

			c := userConstraints[0].(*ReferenceConstraint)
			invProp, _ := c.GetInverseConstrainedPropertyClass()
			g.Assert(invProp).Equal("http://data.mimiro.io/cima/operator")
		})
	})
}

func TestHtmlGenerator(t *testing.T) {
	reader := strings.NewReader(`[
				{ "id" : "@context",
					"namespaces" : {
						"mimiro-schema" : "http://data.mimiro.io/schema/",
						"egcl" : "http://data.mimiro.io/egcl/",
						"rdf" : "http://www.w3.org/1999/02/22-rdf-syntax-ns#"
					}
				}, 

				{
					"id" : "mimiro-schema:Thing",
					"refs" : { "rdf:type" : [ "egcl:EntityClass" ] },
					"props" : {
						"egcl:label" : "Thing",
						"egcl:description" : "Some Thing"
					}
				},
				{
					"id" : "mimiro-schema:OrgUnit",
					"refs" : { 
								"rdf:type" : [ "egcl:EntityClass" ],
								"egcl:subClassOf" : "mimiro-schema:Thing"
							 },
					"props" : {
						"egcl:label" : "Orgnisation Unit",
						"egcl:description" : "A organisation unit, such as a company or department."
					}
				},
				{
					"id" : "mimiro-schema:Person",
					"refs" : { 
								"rdf:type" : [ "egcl:EntityClass" ],
								"egcl:subClassOf" : "mimiro-schema:OrgUnit"
							 },
					"props" : {
						"egcl:label" : "Person",
						"egcl:description" : "A living or dead human"
					}
				},
				{
					"id" : "mimiro-schema:Company",
					"refs" : { "rdf:type" : [ "egcl:EntityClass" ], 
								"egcl:subClassOf" : "mimiro-schema:OrgUnit"
 					},
					"props" : {
						"egcl:label" : "Company",
						"egcl:description" : "A registered organisation allowed to conduct business."
					}
				},
				{
					"id" : "mimiro-schema:constraint-1",
					"refs" : { 
							"rdf:type" : [ "egcl:PropertyConstraint" ],
							"egcl:entityClass" : "mimiro-schema:OrgUnit",
							"egcl:propertyClass" : "mimiro-schema:displayName"	
					},
					"props" : {
							"egcl:minCard" : 0,
							"egcl:maxCard" : 1
					}
				},
				{
					"id" : "mimiro-schema:constraint-2",
					"refs" : { 
							"rdf:type" : [ "egcl:ReferenceConstraint" ],
							"egcl:entityClass" : "mimiro-schema:Person",
							"egcl:referenceClass" : "mimiro-schema:worksFor",
							"egcl:referencedEntityClass" : "mimiro-schema:Company"
					},
					"props" : {
							"egcl:minCard" : 0,
							"egcl:maxCard" : 1
					}
				}]`)

	parser := newParser()
	ec, err := parser.LoadEntityCollection(reader)

	if err != nil {
		t.Fatalf("error parsing schema: " + err.Error())
	}

	NewSchema(ec)

	//	htmlGen := NewHtmlGenerator()
	// 	htmlGen.GenerateHtml(schema, "testschema")
}
