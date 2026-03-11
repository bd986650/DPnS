package org.example;

import org.example.parser.PersonXmlParser;

import java.io.BufferedInputStream;
import java.io.InputStream;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.util.List;

public class Main {
    public static void main(String[] args) throws Exception {
        String input  = args.length > 0 ? args[0] : "src/main/resources/people.xml";
        String output = args.length > 1 ? args[1] : "src/main/resources/output.xml";
        String xsd    = args.length > 2 ? args[2] : "src/main/resources/people.xsd";

        PersonXmlParser parser = new PersonXmlParser();

        try (InputStream in = new BufferedInputStream(Files.newInputStream(Paths.get(input)))) {
            parser.parse(in);
        }

        parser.resolveRelations();
        JaxbWriter.writeWithValidation(parser.people, xsd, output);
    }
}