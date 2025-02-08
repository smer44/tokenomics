# Ren'Py automatically loads all script files ending with .rpy. To use this
# file, define a label and jump to it from another file.


label tests_root:
    menu:
        "test_display_field":
            jump test_display_field 
        "readme_screen_test":
            jump readme_screen_test 

