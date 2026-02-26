package org.example;

import jakarta.xml.bind.annotation.*;
import java.util.ArrayList;
import java.util.List;

@XmlRootElement(name = "people")
@XmlAccessorType(XmlAccessType.FIELD)
public class People {

    @XmlAttribute(name = "count", required = true)
    private int count;

    @XmlElement(name = "person")
    private List<PersonJaxb> persons = new ArrayList<>();

    public People() {}

    public People(List<PersonJaxb> persons) {
        this.persons = persons;
        this.count = persons.size();
    }

    public int getCount() { return count; }
    public List<PersonJaxb> getPersons() { return persons; }
}
