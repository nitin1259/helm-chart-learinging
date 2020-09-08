# Safely open your file
with open('./old.txt', 'r') as f:
   lines = f.readlines()

# Rearrange the lines according to the logic you are looking for
even_lines, odd_lines = lines[::2], lines[1::2]

str1=""
# Safely open your files
with open('sample.txt','w') as f1:
   for odd_line in odd_lines:
       str1 = str1 + odd_line.replace('\n', '').replace('\t','').replace('\r','') + " "
       if odd_line.replace('\n', '').replace('\t','').replace('\r','') == "you will do with secrets manager thank":
            f1.write(str1)

# with open('file2.txt','w') as f2:
#    for even_line in even_lines:
#        f2.write(even_line)



