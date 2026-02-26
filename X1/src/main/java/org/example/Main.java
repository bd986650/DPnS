package org.example;

import java.nio.file.Paths;
import java.util.List;
import java.util.Map;

public class Main {
    public static void main(String[] args) {
        String inputPath  = "src/main/resources/people.xml";
        String outputPath = "src/main/resources/output.xml";
        PersonXmlParser parser = new PersonXmlParser();

        try {
            parser.parse(inputPath);
            Map<String, Person> people = parser.people;
            parser.resolveSiblings();

            // Запись выходного XML
            PersonXmlWriter writer = new PersonXmlWriter(people);
            writer.write(outputPath);
        } catch (Exception e) {
            e.printStackTrace();
        }
    }
}