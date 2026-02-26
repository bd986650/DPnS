package org.example.parser;

import java.util.*;
import java.util.function.Supplier;

public class Person {
    public String id;
    public String firstName;
    public String lastName;
    public String gender;
    public String spouseId;
    public String spouseName;
    public String fatherId, fatherName;
    public String motherId, motherName;
    public List<String> childrenIds, childrenNames;
    public List<String> brotherIds, brotherNames;
    public List<String> sisterIds, sisterNames;
    public List<String> siblingIds, siblingNames;
    public int declaredChildrenCount = -1;

    public List<String> childrenIds() { if (childrenIds == null) childrenIds = new ArrayList<>(2); return childrenIds; }
    public List<String> childrenNames() { if (childrenNames == null) childrenNames = new ArrayList<>(2); return childrenNames; }
    public List<String> brotherIds() { if (brotherIds == null) brotherIds = new ArrayList<>(2); return brotherIds; }
    public List<String> brotherNames() { if (brotherNames == null) brotherNames = new ArrayList<>(2); return brotherNames; }
    public List<String> sisterIds() { if (sisterIds == null) sisterIds = new ArrayList<>(2); return sisterIds; }
    public List<String> sisterNames() { if (sisterNames == null) sisterNames = new ArrayList<>(2); return sisterNames; }
    public List<String> siblingIds() { if (siblingIds == null) siblingIds = new ArrayList<>(2); return siblingIds; }
    public List<String> siblingNames() { if (siblingNames == null) siblingNames = new ArrayList<>(2); return siblingNames; }

    public void addUnique(List<String> list, String val) {
        if (val != null && !list.contains(val)) list.add(val);
    }

    public String fullName() {
        if (firstName != null && lastName != null) return firstName + " " + lastName;
        if (firstName != null) return firstName;
        return lastName;
    }

    public void merge(Person o) {
        if (o.firstName != null) firstName = o.firstName;
        if (o.lastName != null) lastName = o.lastName;
        if (o.gender != null) gender = o.gender;
        if (o.spouseId != null) spouseId = o.spouseId;
        if (o.spouseName != null) spouseName = o.spouseName;
        if (o.fatherId != null) fatherId = o.fatherId;
        if (o.fatherName != null) fatherName = o.fatherName;
        if (o.motherId != null) motherId = o.motherId;
        if (o.motherName != null) motherName = o.motherName;
        if (o.declaredChildrenCount >= 0) declaredChildrenCount = o.declaredChildrenCount;
        mergeLists(this::childrenIds, o.childrenIds);
        mergeLists(this::childrenNames, o.childrenNames);
        mergeLists(this::brotherIds, o.brotherIds);
        mergeLists(this::brotherNames, o.brotherNames);
        mergeLists(this::sisterIds, o.sisterIds);
        mergeLists(this::sisterNames, o.sisterNames);
        mergeLists(this::siblingIds, o.siblingIds);
        mergeLists(this::siblingNames, o.siblingNames);
    }

    private void mergeLists(Supplier<List<String>> target, List<String> source) {
        if (source == null) return;
        List<String> t = target.get();
        for (String s : source) if (!t.contains(s)) t.add(s);
    }
}