package org.example;

import jakarta.xml.bind.annotation.*;

@XmlAccessorType(XmlAccessType.FIELD)
public class Brother {
    @XmlAttribute(name = "id") private String id;
    @XmlValue private String name;
    public Brother() {}
    public Brother(String id, String name) { this.id = id; this.name = name; }
    public String getId() { return id; }
    public String getName() { return name; }
}