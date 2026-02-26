package org.example;

import jakarta.xml.bind.annotation.*;
import java.util.ArrayList;
import java.util.List;

@XmlAccessorType(XmlAccessType.FIELD)
public class Children {

    @XmlAttribute(name = "count", required = true)
    private int count;

    @XmlElements({
            @XmlElement(name = "son",      type = Son.class),
            @XmlElement(name = "daughter",  type = Daughter.class),
            @XmlElement(name = "child",     type = Child.class)
    })
    private List<Object> items = new ArrayList<>();

    public Children() {}

    public int getCount() { return count; }
    public void setCount(int count) { this.count = count; }
    public List<Object> getItems() { return items; }
}