package org.example;

import jakarta.xml.bind.annotation.*;

@XmlAccessorType(XmlAccessType.FIELD)
public class Parents {

    @XmlElement(name = "father")
    private PersonRef father;

    @XmlElement(name = "mother")
    private PersonRef mother;

    public Parents() {}

    public PersonRef getFather() { return father; }
    public void setFather(PersonRef father) { this.father = father; }
    public PersonRef getMother() { return mother; }
    public void setMother(PersonRef mother) { this.mother = mother; }
}