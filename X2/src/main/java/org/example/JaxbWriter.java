package org.example;

import jakarta.xml.bind.*;
import org.example.parser.Person;

import javax.xml.XMLConstants;
import javax.xml.validation.*;
import java.io.*;
import java.nio.file.*;
import java.util.*;

public class JaxbWriter {

    public static void writeWithValidation(
            Map<String, Person> people,
            String xsdPath,
            String outputPath
    ) throws Exception {
        Set<String> allOutputIds = people.keySet();
        List<PersonJaxb> jaxbPersons = new ArrayList<>(people.size());

        for (Person src : people.values()) {
            PersonJaxb jp = new PersonJaxb();
            jp.setId(src.id);
            jp.setFirstName(src.firstName);
            jp.setLastName(src.lastName);

            if (src.gender != null) {
                jp.setGender("M".equals(src.gender) ? "male" : "female");
            }

            // spouse
            if (src.spouseId != null || src.spouseName != null) {
                String sName = src.spouseName != null ? src.spouseName : nameOf(people, src.spouseId);
                PersonRef ref = new PersonRef();
                ref.setId(safeId(src.spouseId, allOutputIds));
                ref.setName(sName);
                jp.setSpouse(ref);
            }

            // parents
            Parents par = new Parents();
            boolean hasParents = false;
            if (src.fatherId != null || src.fatherName != null) {
                String fn = src.fatherName != null ? src.fatherName : nameOf(people, src.fatherId);
                PersonRef ref = new PersonRef();
                ref.setId(safeId(src.fatherId, allOutputIds));
                ref.setName(fn);
                par.setFather(ref);
                hasParents = true;
            }
            if (src.motherId != null || src.motherName != null) {
                String mn = src.motherName != null ? src.motherName : nameOf(people, src.motherId);
                PersonRef ref = new PersonRef();
                ref.setId(safeId(src.motherId, allOutputIds));
                ref.setName(mn);
                par.setMother(ref);
                hasParents = true;
            }
            if (hasParents) jp.setParents(par);

            // children
            int childCount = sizeCol(src.childrenIds) + sizeCol(src.childrenNames);
            if (childCount > 0) {
                Children ch = new Children();
                if (src.childrenIds != null) {
                    for (String cid : src.childrenIds) {
                        Person cp = people.get(cid);
                        String cn = cp != null ? getFullName(cp) : null;
                        String g = cp != null ? cp.gender : null;
                        String safeCid = safeId(cid, allOutputIds);
                        if ("M".equals(g))       ch.getItems().add(new Son(safeCid, cn));
                        else if ("F".equals(g))   ch.getItems().add(new Daughter(safeCid, cn));
                        else                       ch.getItems().add(new Child(safeCid, cn));
                    }
                }
                if (src.childrenNames != null) {
                    for (String cn : src.childrenNames) {
                        ch.getItems().add(new Child(null, cn));
                    }
                }
                ch.setCount(childCount);
                jp.setChildren(ch);
            }

            // siblings
            boolean hasSib = sizeCol(src.brotherIds) + sizeCol(src.brotherNames)
                    + sizeCol(src.sisterIds) + sizeCol(src.sisterNames)
                    + sizeCol(src.siblingIds) + sizeCol(src.siblingNames) > 0;
            if (hasSib) {
                Siblings sib = new Siblings();
                addItems(sib, src.brotherIds, people, "brother", allOutputIds);
                addNameItems(sib, src.brotherNames, "brother");
                addItems(sib, src.sisterIds, people, "sister", allOutputIds);
                addNameItems(sib, src.sisterNames, "sister");
                addItems(sib, src.siblingIds, people, "sibling", allOutputIds);
                addNameItems(sib, src.siblingNames, "sibling");
                jp.setSiblings(sib);
            }

            jaxbPersons.add(jp);
        }

        People root = new People(jaxbPersons);

        // Schema
        SchemaFactory sf = SchemaFactory.newInstance(XMLConstants.W3C_XML_SCHEMA_NS_URI);
        Schema schema = sf.newSchema(Paths.get(xsdPath).toFile());

        // Marshal
        JAXBContext ctx = JAXBContext.newInstance(People.class);
        Marshaller m = ctx.createMarshaller();
        m.setProperty(Marshaller.JAXB_FORMATTED_OUTPUT, true);
        m.setProperty(Marshaller.JAXB_ENCODING, "UTF-8");
        m.setSchema(schema);

        m.setEventHandler(event -> {
            System.err.println("Validation: " + event.getMessage());
            return true; // продолжаем
        });

        Path out = Paths.get(outputPath);
        if (out.getParent() != null) Files.createDirectories(out.getParent());

        try (OutputStream os = new BufferedOutputStream(Files.newOutputStream(out), 1 << 16)) {
            m.marshal(root, os);
        }

        System.out.println("JAXB output → " + out.toAbsolutePath());
    }

    private static String safeId(String id, Set<String> allOutputIds) {
        return (id != null && allOutputIds.contains(id)) ? id : null;
    }

    private static void addItems(Siblings sib, Collection<String> ids,
                                 Map<String, Person> people,
                                 String type, Set<String> allOutputIds) {
        if (ids == null) return;
        for (String id : ids) {
            Person p = people.get(id);
            String name = p != null ? getFullName(p) : null;
            sib.getItems().add(createByType(type, safeId(id, allOutputIds), name));
        }
    }

    private static void addNameItems(Siblings sib, Collection<String> names, String type) {
        if (names == null) return;
        for (String n : names) {
            sib.getItems().add(createByType(type, null, n));
        }
    }

    private static Object createByType(String type, String id, String name) {
        switch (type) {
            case "brother": return new Brother(id, name);
            case "sister":  return new Sister(id, name);
            default:        return new SiblingRef(id, name);
        }
    }

    private static int sizeCol(Collection<?> c) {
        return c == null ? 0 : c.size();
    }

    private static String getFullName(Person p) {
        if (p.firstName != null && p.lastName != null) return p.firstName + " " + p.lastName;
        if (p.firstName != null) return p.firstName;
        return p.lastName;
    }

    private static String nameOf(Map<String, Person> people, String id) {
        if (id == null) return null;
        Person p = people.get(id);
        return p != null ? getFullName(p) : null;
    }
}