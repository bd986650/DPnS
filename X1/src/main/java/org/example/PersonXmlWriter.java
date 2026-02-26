package org.example;

import java.io.*;
import java.nio.file.*;
import java.util.List;
import java.util.Map;

import static org.example.XmlUtils.esc;

public class PersonXmlWriter {

    private final Map<String, Person> people;

    public PersonXmlWriter(Map<String, Person> people) {
        this.people = people;
    }

    public void write(String outputPath) throws IOException {
        Path out = Paths.get(outputPath);
        Files.createDirectories(out.getParent());

        try (BufferedWriter bw = Files.newBufferedWriter(out, java.nio.charset.StandardCharsets.UTF_8)) {
            bw.write("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n");
            bw.write("<people count=\"" + people.size() + "\">\n");

            for (Person p : people.values()) {
                bw.write("  <person id=\"" + esc(p.id) + "\">\n");

                // Основные поля
                writeTag(bw, "first-name", p.firstName);
                writeTag(bw, "last-name", p.lastName);
                if (p.gender != null) writeTag(bw, "gender", "M".equals(p.gender) ? "male" : "female");

                // супруг/супруга
                if (p.spouseId != null || p.spouseName != null) {
                    String st = "M".equals(p.gender) ? "wife" : "F".equals(p.gender) ? "husband" : "spouse";
                    bw.write("    <" + st);
                    if (p.spouseId != null) bw.write(" id=\"" + p.spouseId + "\"");
                    String sn = p.spouseName != null ? p.spouseName : nameOf(p.spouseId);
                    if (sn != null) {
                        bw.write(">" + esc(sn) + "</" + st + ">\n");
                    } else {
                        bw.write("/>\n");
                    }
                }

                // родители
                bw.write("    <parents>\n");
                writeRef(bw, "father", p.fatherId, p.fatherName);
                writeRef(bw, "mother", p.motherId, p.motherName);
                bw.write("    </parents>\n");

                // дети
                int childCount = size(p.childrenIds) + size(p.childrenNames);
                if (childCount > 0) {
                    bw.write("    <children count=\"" + childCount + "\">\n");
                    if (p.childrenIds != null) {
                        for (String cid : p.childrenIds) {
                            Person ch = people.get(cid);
                            String ct = ch != null && "M".equals(ch.gender) ? "son" :
                                    ch != null && "F".equals(ch.gender) ? "daughter" : "child";
                            writeRef(bw, ct, cid, ch != null ? ch.fullName() : null);
                        }
                    }
                    if (p.childrenNames != null) {
                        for (String cn : p.childrenNames) writeRef(bw, "child", null, cn);
                    }
                    bw.write("    </children>\n");
                }

                // сиблинги
                boolean hasSib = size(p.brotherIds) + size(p.brotherNames)
                        + size(p.sisterIds) + size(p.sisterNames)
                        + size(p.siblingIds) + size(p.siblingNames) > 0;
                if (hasSib) {
                    bw.write("    <siblings>\n");
                    writeRefs(bw, "brother", p.brotherIds, p.brotherNames);
                    writeRefs(bw, "sister", p.sisterIds, p.sisterNames);
                    writeRefs(bw, "sibling", p.siblingIds, p.siblingNames);
                    bw.write("    </siblings>\n");
                }

                bw.write("  </person>\n");
            }

            bw.write("</people>\n");
        }
    }

    private void writeTag(BufferedWriter bw, String tag, String val) throws IOException {
        if (val == null) return;
        bw.write("    <" + tag + ">" + esc(val) + "</" + tag + ">\n");
    }

    private void writeRef(BufferedWriter bw, String tag, String id, String name) throws IOException {
        if (id == null && name == null) return;
        bw.write("      <" + tag);
        if (id != null) bw.write(" id=\"" + id + "\"");
        if (name != null) bw.write(">" + esc(name) + "</" + tag + ">\n");
        else bw.write("/>\n");
    }

    private void writeRefs(BufferedWriter bw, String tag, List<String> ids, List<String> names) throws IOException {
        if (ids != null) for (String id : ids) {
            Person ref = people.get(id);
            writeRef(bw, tag, id, ref != null ? ref.fullName() : null);
        }
        if (names != null) for (String n : names) writeRef(bw, tag, null, n);
    }

    private String nameOf(String id) {
        if (id == null) return null;
        Person p = people.get(id);
        return p != null ? p.fullName() : null;
    }

    private static int size(List<?> l) { return l == null ? 0 : l.size(); }
}