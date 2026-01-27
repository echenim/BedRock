public Map<MemberUrn, MemberColoAssignment> batchGetAssignments(final Set<MemberUrn> members) {
    final Map<MemberUrn, MemberColoAssignment> map = new HashMap<>();

    for (final MemberUrn member : members) {
      final long memberId = member.getMemberIdEntity();

      if (memberId > 0 && !map.containsKey(member)) {
        final MemberColoAssignment assignment = getColoAssignment(member);
        map.put(member, assignment);
      }
    }
    return map;
  }
