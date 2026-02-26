package org.example;

import jakarta.xml.bind.annotation.*;

@XmlAccessorType(XmlAccessType.FIELD)
public class PersonJaxb {

    @XmlAttribute(name = "id", required = true)
    @XmlID
    private String id;

    @XmlElement(name = "first-name")
    private String firstName;

    @XmlElement(name = "last-name")
    private String lastName;

    @XmlElement(name = "gender")
    private String gender;

    @XmlElement(name = "spouse")
    private PersonRef spouse;

    @XmlElement(name = "parents")
    private Parents parents;

    @XmlElement(name = "children")
    private Children children;

    @XmlElement(name = "siblings")
    private Siblings siblings;

    public PersonJaxb() {}

    public String getId() { return id; }
    public void setId(String id) { this.id = id; }
    public String getFirstName() { return firstName; }
    public void setFirstName(String firstName) { this.firstName = firstName; }
    public String getLastName() { return lastName; }
    public void setLastName(String lastName) { this.lastName = lastName; }
    public String getGender() { return gender; }
    public void setGender(String gender) { this.gender = gender; }
    public PersonRef getSpouse() { return spouse; }
    public void setSpouse(PersonRef spouse) { this.spouse = spouse; }
    public Parents getParents() { return parents; }
    public void setParents(Parents parents) { this.parents = parents; }
    public Children getChildren() { return children; }
    public void setChildren(Children children) { this.children = children; }
    public Siblings getSiblings() { return siblings; }
    public void setSiblings(Siblings siblings) { this.siblings = siblings; }
}