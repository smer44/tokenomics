# Ren'Py automatically loads all script files ending with .rpy. To use this
# file, define a label and jump to it from another file.

init -1000 python:

    #readme_path = "../README.md"
    #readme_path = "../test.txt"



    

    def count_characters(s, target):
        counter = 0
        for i in range(len(s)):
            if s[i] == target:
                counter += 1
            else:
                return counter

    def tag_str(tag,txt):
        closed_tag = tag.split("=",1)[0]
        return "{" + tag  + "}" + txt + "{/" + closed_tag + "}"
        #return f"\{{tag}\}\{{txt}\}\{/{closed_tag}\}"

    md_tags_dict = {"**": "b"}
                    

    def apply_md_brackets(line, br):
        tag = md_tags_dict[br]
        lines = line.split(br)
        lines = [tag_str(tag,word) if n%2 else word for n,word in enumerate(lines)]
        return "".join(lines)


    def apply_md_string(line):
        '''
        now, only ** and # are supported 
        '''
        line = line .replace("\*","*")
        header_level = count_characters(line,"#")
        if header_level:
            font_size_add = (7 - header_level) *3
            tag = f"size=+{font_size_add}"
            line_after_header = line[header_level:]
            line_after_header = apply_md_brackets(line_after_header, "**")
            return tag_str(tag,line_after_header)
        else:
            return apply_md_brackets(line, "**")
        

    readme_file = renpy.open_file("../README.md", "utf-8")   

    readme_lines = [apply_md_string(e) for e in readme_file.readlines()]    



    game_rules_file =  renpy.open_file("../docs/game-rules.md", "utf-8")

    game_rules_lines = [apply_md_string(e) for e in game_rules_file.readlines()]    
        






screen markdown_lines(lines):
    viewport:
        mousewheel True 
        draggable True 
        vbox:
            for line in lines:
                text "[line]" 



label readme_screen:
    show screen markdown_lines(readme_lines)
    jump lloop 


label game_rules_screen:
    show screen markdown_lines(game_rules_lines)
    jump lloop 


label documents:
    menu:
        "Readme":
            jump readme_screen
        "Game rules":
            jump game_rules_screen

