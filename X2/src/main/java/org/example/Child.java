package org.example;

import jakarta.xml.bind.annotation.*;

@XmlAccessorType(XmlAccessType.FIELD)
public class Child {
    @XmlAttribute(name = "id") private String id;
    @XmlValue private String name;
    public Child() {}
    public Child(String id, String name) { this.id = id; this.name = name; }
    public String getId() { return id; }
    public String getName() { return name; }
}