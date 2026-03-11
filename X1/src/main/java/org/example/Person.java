package org.example;

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

    public void merge(Person p) {
        if (p.firstName != null) firstName = p.firstName;
        if (p.lastName != null) lastName = p.lastName;
        if (p.gender != null) gender = p.gender;
        if (p.spouseId != null) spouseId = p.spouseId;
        if (p.spouseName != null) spouseName = p.spouseName;
        if (p.fatherId != null) fatherId = p.fatherId;
        if (p.fatherName != null) fatherName = p.fatherName;
        if (p.motherId != null) motherId = p.motherId;
        if (p.motherName != null) motherName = p.motherName;
        if (p.declaredChildrenCount >= 0) declaredChildrenCount = p.declaredChildrenCount;
        mergeLists(this::childrenIds, p.childrenIds);
        mergeLists(this::childrenNames, p.childrenNames);
        mergeLists(this::brotherIds, p.brotherIds);
        mergeLists(this::brotherNames, p.brotherNames);
        mergeLists(this::sisterIds, p.sisterIds);
        mergeLists(this::sisterNames, p.sisterNames);
        mergeLists(this::siblingIds, p.siblingIds);
        mergeLists(this::siblingNames, p.siblingNames);
    }

    private void mergeLists(Supplier<List<String>> target, List<String> source) {
        if (source == null) return;
        List<String> t = target.get();
        for (String s : source) if (!t.contains(s)) t.add(s);
    }
}