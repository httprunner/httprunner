# Comparator

| comparator | Description | A(check), B(expected) | examples |
|:------------:|-----------|------------------------------------|-----------|
| `eq`, `==` | value is equal | A == B | 9 eq 9 |
| `lt` | less than | A < B | 7 lt 8 |
| `le` | less than or equals | A <= B | 7 le 8, 8 le 8 |
| `gt` | greater than | A > B | 8 gt 7 |
| `ge` | greater than or equals | A >= B | 8 ge 7, 8 ge 8 |
| `ne` | not equals | A != B | 6 ne 9 |
| `str_eq` | string equals | str(A) == str(B) | 123 str_eq '123' |
| `len_eq`, `count_eq` | length or count equals | len(A) == B | 'abc' len_eq 3<br/>[1,2] len_eq 2 |
| `len_gt`, `count_gt` | length greater than | len(A) > B | 'abc' len_gt 2<br/>[1,2,3] len_gt 2 |
| `len_ge`, `count_ge` | length greater than or equals | len(A) >= B | 'abc' len_ge 3<br/>[1,2,3] len_gt 3 |
| `len_lt`, `count_lt` | length less than | len(A) < B | 'abc' len_lt 4<br/>[1,2,3] len_lt 4 |
| `len_le`, `count_le` | length less than or equals | len(A) <= B | 'abc' len_le 3<br/>[1,2,3] len_le 3 |
| `contains` | contains | B in A | [1, 2] contains 1<br/>'abc' contains 'a' |
| `contained_by` | contained by | A in B | 1 contained_by [1,2]<br/>'a' contained_by 'abc' |
| `type` | type of A is instance of B | isinstance(A, B) | 123 type 'int' |
| `regex` | regex matches | re.match(B, A) | 'abcdef' regex 'a\w+d' |
| `startswith` | starts with | A.startswith(B) is True | 'abc' startswith 'ab' |
| `endswith` | ends with | A.endswith(B) is True | 'abc' endswith 'bc' |
