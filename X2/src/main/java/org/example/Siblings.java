package org.example;

import jakarta.xml.bind.annotation.*;
import java.util.ArrayList;
import java.util.List;

@XmlAccessorType(XmlAccessType.FIELD)
public class Siblings {

    @XmlElements({
            @XmlElement(name = "brother", type = Brother.class),
            @XmlElement(name = "sister",  type = Sister.class),
            @XmlElement(name = "sibling", type = SiblingRef.class)
    })
    private List<Object> items = new ArrayList<>();

    public Siblings() {}

    public List<Object> getItems() { return items; }
}