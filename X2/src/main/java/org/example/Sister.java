package org.example;

import jakarta.xml.bind.annotation.*;

@XmlAccessorType(XmlAccessType.FIELD)
public class Sister {
    @XmlAttribute(name = "id") private String id;
    @XmlValue private String name;
    public Sister() {}
    public Sister(String id, String name) { this.id = id; this.name = name; }
    public String getId() { return id; }
    public String getName() { return name; }
}