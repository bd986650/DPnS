package org.example;

import jakarta.xml.bind.annotation.*;

@XmlAccessorType(XmlAccessType.FIELD)
public class Son {
    @XmlAttribute(name = "id") private String id;
    @XmlValue private String name;
    public Son() {}
    public Son(String id, String name) { this.id = id; this.name = name; }
    public String getId() { return id; }
    public String getName() { return name; }
}