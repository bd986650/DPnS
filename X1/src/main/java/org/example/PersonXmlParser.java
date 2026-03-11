package org.example;

import javax.xml.stream.*;
import java.io.*;
import java.nio.file.*;
import java.util.*;

public class PersonXmlParser {
    public final Map<String, Person> people = new HashMap<>(40000);
    private int declaredPeopleCount = -1;

    public void parse(String inputPath) throws Exception {
        XMLInputFactory f = XMLInputFactory.newInstance();
        f.setProperty(XMLInputFactory.IS_COALESCING, false);
        f.setProperty(XMLInputFactory.SUPPORT_DTD, false);
        f.setProperty(XMLInputFactory.IS_SUPPORTING_EXTERNAL_ENTITIES, false);
        f.setProperty(XMLInputFactory.IS_REPLACING_ENTITY_REFERENCES, false);

        try (InputStream raw = new BufferedInputStream(Files.newInputStream(Paths.get(inputPath)), 1 << 17)) {
            XMLStreamReader r = f.createXMLStreamReader(raw, "UTF-8");
            Person cur = null;
            String curId = null;
            String elem = null;
            String attrVal = null, attrCount = null;
            StringBuilder text = new StringBuilder(128);

            while (r.hasNext()) {
                switch (r.next()) {
                    case XMLStreamConstants.START_ELEMENT:
                        elem = r.getLocalName().toLowerCase();
                        text.setLength(0);
                        attrVal = null; attrCount = null;
                        for (int i = 0, n = r.getAttributeCount(); i < n; i++) {
                            String an = r.getAttributeLocalName(i).toLowerCase();
                            String av = r.getAttributeValue(i).trim();
                            if (an.equals("value") || an.equals("val") || an.equals("id")) attrVal = av;
                            else if (an.equals("count")) attrCount = av;
                        }
                        if ("people".equals(elem) && attrCount != null) {
                            try { declaredPeopleCount = Integer.parseInt(attrCount); } catch (NumberFormatException ignored) {}
                        } else if ("person".equals(elem)) {
                            curId = attrVal;
                            cur = new Person(); cur.id = curId;
                        } else if (cur != null) applyAttr(elem, attrVal, attrCount, cur);
                        break;

                    case XMLStreamConstants.CHARACTERS:
                    case XMLStreamConstants.CDATA:
                        if (cur != null && elem != null) {
                            int len = r.getTextLength();
                            if (len > 0) text.append(r.getTextCharacters(), r.getTextStart(), len);
                        }
                        break;

                    case XMLStreamConstants.END_ELEMENT:
                        String endName = r.getLocalName().toLowerCase();
                        if ("person".equals(endName)) {
                            if (cur != null && curId != null) {
                                Person existing = people.get(curId);
                                if (existing == null) people.put(curId, cur);
                                else existing.merge(cur);
                            }
                            cur = null; curId = null; elem = null;
                        } else if (cur != null) {
                            String t = text.toString().trim();
                            if (!t.isEmpty()) applyText(endName, t, cur);
                            text.setLength(0); elem = null;
                        }
                        break;
                }
            }
            r.close();
        }
    }

    private static boolean isId(String s) { return s != null && s.length() > 1 && s.charAt(0) == 'P' && Character.isDigit(s.charAt(1)); }

    private static String normGender(String g) {
        char c = Character.toLowerCase(g.charAt(0));
        return (c == 'm') ? "M" : (c == 'f' || c == 'w') ? "F" : null;
    }

    private static void parseName(String n, Person p) {
        String[] parts = n.split("\\s+", 2);
        if (p.firstName == null && parts.length >= 1) p.firstName = parts[0];
        if (p.lastName == null && parts.length == 2) p.lastName = parts[1];
    }

    private void applyAttr(String tag, String val, String count, Person p) {
        if (val == null && count == null) return;
        switch (tag) {
            case "firstname": case "first-name": case "first_name":
                if (val != null) p.firstName = val; break;
            case "lastname": case "last-name": case "last_name": case "surname":
                if (val != null) p.lastName = val; break;
            case "name": case "fullname": case "full-name":
                if (val != null) parseName(val, p); break;
            case "gender": case "sex":
                if (val != null) p.gender = normGender(val); break;

            case "husband":
                if (p.gender == null) p.gender = "F";
                if (isId(val)) p.spouseId = val; else if (val != null) p.spouseName = val; break;
            case "wife":
                if (p.gender == null) p.gender = "M";
                if (isId(val)) p.spouseId = val; else if (val != null) p.spouseName = val; break;
            case "spouse":
                if (isId(val)) p.spouseId = val; else if (val != null) p.spouseName = val; break;

            case "father":
                if (isId(val)) p.fatherId = val; else if (val != null) p.fatherName = val; break;
            case "mother":
                if (isId(val)) p.motherId = val; else if (val != null) p.motherName = val; break;
            case "parent":
                if (isId(val)) { if (p.fatherId == null) p.fatherId = val; else p.motherId = val; }
                break;

            case "child": case "son": case "daughter":
                if (isId(val)) p.addUnique(p.childrenIds(), val); else if (val != null) p.addUnique(p.childrenNames(), val); break;
            case "children": case "children-number": case "childrencount": case "children-count":
                String c = count != null ? count : val;
                if (c != null) try { p.declaredChildrenCount = Integer.parseInt(c); } catch (NumberFormatException ignored) {} break;

            case "siblings": case "sibling":
                if (val != null) {
                    for (String id : val.split("\\s+")) {
                        if (isId(id)) p.addUnique(p.siblingIds(), id);
                        else p.addUnique(p.siblingNames(), id);
                    }
                }
                break;
            case "brother":
                if (isId(val)) p.addUnique(p.brotherIds(), val); else if (val != null) p.addUnique(p.brotherNames(), val); break;
            case "sister":
                if (isId(val)) p.addUnique(p.sisterIds(), val); else if (val != null) p.addUnique(p.sisterNames(), val); break;
        }
    }

    private void applyText(String tag, String t, Person p) {
        switch (tag) {
            case "firstname": case "first-name": case "first_name":
                p.firstName = t; break;
            case "lastname": case "last-name": case "last_name": case "surname":
                p.lastName = t; break;
            case "name": case "fullname": case "full-name":
                parseName(t, p); break;
            case "gender": case "sex":
                p.gender = normGender(t); break;
            case "husband":
                if (p.gender == null) p.gender = "F";
                if (isId(t)) p.spouseId = t; else p.spouseName = t; break;
            case "wife":
                if (p.gender == null) p.gender = "M";
                if (isId(t)) p.spouseId = t; else p.spouseName = t; break;
            case "spouse":
                if (isId(t)) p.spouseId = t; else p.spouseName = t; break;
            case "father":
                if (isId(t)) p.fatherId = t; else p.fatherName = t; break;
            case "mother":
                if (isId(t)) p.motherId = t; else p.motherName = t; break;
            case "parent":
                if (isId(t)) { if (p.fatherId == null) p.fatherId = t; else p.motherId = t; }
                else { if (p.fatherName == null) p.fatherName = t; else p.motherName = t; } break;
            case "child": case "son": case "daughter":
                if (isId(t)) p.addUnique(p.childrenIds(), t); else p.addUnique(p.childrenNames(), t); break;
            case "children": case "children-number": case "childrencount": case "children-count":
                try { p.declaredChildrenCount = Integer.parseInt(t); } catch (NumberFormatException ignored) {} break;
            case "siblings": case "sibling":
                if (isId(t)) p.addUnique(p.siblingIds(), t);
                else p.addUnique(p.siblingNames(), t);
                break;
            case "brother":
                if (isId(t)) p.addUnique(p.brotherIds(), t); else p.addUnique(p.brotherNames(), t); break;
            case "sister":
                if (isId(t)) p.addUnique(p.sisterIds(), t); else p.addUnique(p.sisterNames(), t); break;
        }
    }

    public void resolveRelations() {
        for (Person p : people.values()) {
            if (p.fatherId != null) {
                Person father = people.get(p.fatherId);
                if (father != null && father.gender == null) {
                    father.gender = "M";
                }
            }

            if (p.motherId != null) {
                Person mother = people.get(p.motherId);
                if (mother != null && mother.gender == null) {
                    mother.gender = "F";
                }
            }

            // <parent value="P10"/>   <parent value="P20"/>
            // fatherId = P10           motherId = P20
            // P10.gender = "F"         P20.gender = "M"
            if (p.fatherId != null && p.motherId != null) {
                Person f = people.get(p.fatherId);
                if (f != null && "F".equals(f.gender)) {
                    String tmp = p.fatherId;
                    p.fatherId = p.motherId;
                    p.motherId = tmp;
                }
            }

            if (p.spouseId != null) {
                Person spouse = people.get(p.spouseId);
                if (spouse != null) {
                    if (spouse.spouseId == null) spouse.spouseId = p.id;
                    if (p.gender != null && spouse.gender == null) {
                        spouse.gender = "M".equals(p.gender) ? "F" : "M";
                    }
                }
            }

            if (p.childrenIds() != null) {
                for (String cid : p.childrenIds()) {
                    Person child = people.get(cid);
                    if (child == null) continue;
                    if ("M".equals(p.gender)) {
                        if (child.fatherId == null) child.fatherId = p.id;
                    } else if ("F".equals(p.gender)) {
                        if (child.motherId == null) child.motherId = p.id;
                    }
                }
            }

            if (p.fatherId != null) {
                Person father = people.get(p.fatherId);
                if (father != null) father.addUnique(father.childrenIds(), p.id);
            }
            if (p.motherId != null) {
                Person mother = people.get(p.motherId);
                if (mother != null) mother.addUnique(mother.childrenIds(), p.id);
            }

            if (p.siblingIds != null) {
                Iterator<String> it = p.siblingIds.iterator();
                while (it.hasNext()) {
                    String sid = it.next();
                    Person sib = people.get(sid);
                    if (sib == null) continue;
                    if ("M".equals(sib.gender)) p.addUnique(p.brotherIds(), sid); it.remove();
                    if ("F".equals(sib.gender)) p.addUnique(p.sisterIds(), sid); it.remove();
                }
            }
        }
    }
}
