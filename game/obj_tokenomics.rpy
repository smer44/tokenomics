

# Ren'Py automatically loads all script files ending with .rpy. To use this
# file, define a label and jump to it from another file.


init -1000 python:

    from random import random as rnd 
    from random import sample as sample

    class yNeed:

        clz = "Need"

        def __init__(self,wait,punish,reward,amount, max ):
            self.wait = wait 
            self.age = 0
            self.punish =punish
            self.reward = reward 
            self.amount = amount
            self.max = max 
            self.free = True 
            self.fullfilled = False 
            self.human = None
            self.set_random_res(amount, max )


        def set_random_res(self,amount, max ):
            ids = sample(range(max),amount) 
            self.res = [0 for _ in range(max)]
            for id in ids:
                self.res[id] = 1

        

        def tick(self):
            human = self.human 
            self.age += 1
            if self.age >=self.wait:
                human.happy -= self.punish


        def reward(self):
            human = self.human
            human.happy += self.reward
            human.needs.remove(self)


        def clone(self):
            return yNeed(self.wait, self.punish,self.reward, self.amount, self.max)


    simple_need = yNeed(3,1,5,1,4)


    middle_need = yNeed(3,2,8,3,4)


    hard_need = yNeed(3,3,10,4,4)    


    needs = [simple_need, middle_need, hard_need]

        
    class tAgent:

        clz = "Agent"

        images = ["agent_idle.png", "agent_moving.png", "agent_working.png"]

        def __init__(self,x,y):
            self.x = x 
            self.y = y 
            self.tokens = 20 
            self.task_stack = [] 
            self.inventory = [0 for _ in range(4)]
            self.task_step  = None 

        def select_task(self,field):
            needs = field.needs
            new_need = needs[field.new_need]
            field.new_need += 1

            factory = field.clothest(self.x,self.y,"Factory")
            need_res = factory.need_res(new_need,self.inventory)
            if need_res:
                self.task_stack.append(("gather",need_res))

            self.task_stack.append(("move",factory.x, factory.y))
            self.task_stack.append(("produce",factory,new_need))







        def is_moving(self):
            return self.task_step and "move" == self.task_step[0]

        def is_working(self):
            return self.task_step and "work" == self.task_step[0]

        def image(self):
            if self.is_moving():
                return self.images[1]
            if self.is_working():
                return self.images[2]
            return self.images[0]



        def move(self,xx,yy,field):
            x,y = self.x , self.y
            dx = xx -x 
            dy = yy -y 
            d = (dx*dx + dy*dy)**0.5 
            max_step = 3
            if d <=max_step:
                self.x, self.y = xx,yy 
                field.grid[x][y].remove(self)
                field.grid[xx][yy].append(self)
                arrived = True
                renpy.notify(f"{self.x=}, {self.y=}, {d=},  {arrived=}")
                return True 

            step_x = int(round(max_step*dx / d ))
            step_y = int(round(max_step*dy / d ))
            
            self.x += step_x
            self.y += step_y 
            field.grid[x][y].remove(self)
            field.grid[self.x][self.y].append(self)            
            
            arrived = self.x == xx and self.y == yy 
            renpy.notify(f"{self.x=}, {self.y=}, {step_x=}, {step_y=}, {d=}, {arrived=}")
            return arrived

        def set_move_task_step(self,xx,yy):
            self.task_step = ("move",xx,yy)

        def execute_task_step(self,field):
            if self.task_step:
                type, *args = self.task_step
                if type == "move":
                    xx,yy = args
                    arrived = self.move(xx,yy,field)
                    if arrived:
                        self.task_step = None 


    def addv(lst,lst2):
        return [lst[n] + lst2[n] for n in range(len(lst))]

    def subv(lst,lst2):
        return [lst[n] - lst2[n] for n in range(len(lst))]

    def maxvc(lst,c):
        return [max(lst[n],c) for n in range(len(lst))]    

    class tFactory:

        clz = "Factory"

        images = ["factory_idle.png" , "factory_works.png"]

        def __init__(self,x,y):
            self.x = x 
            self.y = y 
            self.res = [5,5,5,5]
            self.free = True 
            self.working = False 

        def image(self):
            return self.images[self.working]


        def state(self):
            return self.working


        def need_res(self,need,inventory):
            all_res = addv(self.res, inventory)
            deficite = maxvc(subv(need.res,all_res),0)
            if any(deficite):
                return deficite
            return None 



    class tRessource:

        clz = "Ressource"

        images = ["res0.png" , "res1.png", "res2.png" , "res3.png"]

        def __init__(self,x,y):
            self.x = x 
            self.y = y
            self.type_id = int(rnd()*3.9)
            self.amount = 1+int(rnd()*1.9) 

        def image(self):
            return self.images[self.type_id]



    class tCell:


        def __init__(self):
            self.humans = []
            self.agents = []
            self.factories = []
            self.resources = []
            self.lists={"Human" : self.humans,
                        "Agent" : self.agents,
                        "Factory": self.factories,
                        "Ressource" : self.resources,
                        }


        def append(self,obj):
            lst = self.lists[obj.clz]
            lst.append(obj)

        def items(self):
            return self.factories + self.humans + self.agents + self.resources

        def remove(self,obj):
            lst = self.lists[obj.clz]
            lst.remove(obj)

        def clear(self):
            for key, lst in self.lists.items():
                lst.clear()
        


    class tField:

        def __init__(self,w,h):
            self.w = w 
            self.h = h
            self.init()
            self.lists={"Human" : self.humans,
                        "Agent" : self.agents,
                        "Factory": self.factories,
                        "Ressource" : self.resources,
                        "Need" : self.needs}


        def clothest(self,x,y,clz_name):
            lst = self.lists[clz_name]
            dmin = 10000000
            selected = None 
            for n, item in enumerate(lst):
                xx,yy = item.x, item.y 
                d = abs(xx-x) + abs(yy - y)
                if d < dmmin:
                    dmin = d 
                    selected = item 
            return selected




        def shape(self):
            return self.w , self.h 


        def random_cell(self):
            return int(rnd()*self.w) , int(rnd() * self.h)
        
        def new_random_obj(self,clz):
            x,y = self.random_cell()
            obj = clz(x,y )
            self.grid[x][y].append(obj)
            return obj 


        def init(self):
            w,h = self.w, self.h
            self.needs = []
            self.new_need = 0
            self.humans = []
            self.agents = []
            self.factories = []
            self.resources = []
            self.grid = [[tCell() for y in range(h)] for x in range(w)]



        def clear(self):

            self.needs.clear()
            self.humans.clear()
            self.agents.clear()
            self.factories.clear()
            self.resources.clear()
            for row in self.grid:
                for cell in row:
                    cell.clear()


        def reset(self):
            self.clear()


            for _ in range(5):
                #random factories:
                self.new_random_obj(tFactory)
            for _ in range(5):
                #random humans:
                human = self.new_random_obj(tHuman)
                self.humans.append(human)
            for _ in range(1):
                agent = self.new_random_obj(tAgent)
                self.agents.append(agent)
                #set random movement for test:
                destination = self.random_cell()
                agent.set_move_task_step(*destination)

            for _ in range(5):
                ressource = self.new_random_obj(tRessource)
                self.resources.append(ressource)

        def tick_all(self):
            self.tick_humans()
            self.tick_agents()
            
        def tick_humans(self):
            for h in self.humans:
                h.tick(self)

        def tick_agents(self):
            for a in self.agents:
                a.execute_task_step(self)

        



    class tHuman:
        

        clz = "Human"

        images = ["human_awsome.png", "human_ok.png", "human_warning.png" , "human_critical.png"]

        def __init__(self,x,y):
            self.x = x 
            self.y = y 
            self.happy = 50 
            self.needs = []
            self.tokens = 0

        def state(self):
            return self.happy


        def image(self):
            if self.happy >= 70:
                return self.images[0]
            if self.is_critical():
                return self.images[3]
            if self.is_warning():
                return self.images[2]
            return self.images[1]


        def is_warning(self):
            return self.happy < 20 

        def is_critical(self):
            return self.happy < 1 


        def tick(self,field):
            new_need = self.new_random_need()
            
            if new_need :
                new_need.human = self 
                new_need.x = self.x 
                new_need.y = self.y 
                self.needs.append(new_need)
                field.needs.append(new_need)

            self.punish_all_needs()
            self.set_happy(self.happy)
            self.gen_tokens()



        def gen_tokens(self):
            if self.happy > 20:
                tokens = 5 + int(rnd() * 5)
            elif self.happy > 0:
                tokens = 3 + int(rnd() * 3)
            else:
                tokens = 0
            self.tokens+=tokens




        def punish_all_needs(self):
            for need in self.needs:
                need.tick(self)
            




        def set_happy(self,value):
            self.happy = min(100, max(0, value))


        def new_random_need(self):
            choice = int(rnd() * 10)
            if choice < 3:
                return needs[choice].clone()
            return None 

        def __str__(self):
            return f"Hu({self.happy},{self.tokens})"

        def __repr__(self):
            return f"Hu({self.happy},{self.tokens})"


screen display_field(field):
    default cell_size = 70

    vbox:
        text "[field.humans!q]"
        align(0.5,0.5)
        textbutton "reset" action Function(field.reset)
        textbutton "tick_all" action Function(field.tick_all)
        grid field.h field.w:
            spacing 3 
            for x in range(field.w):
                for y in range(field.h):
                    $ cell = field.grid[x][y]
                    frame:
                        style "empty"
                        xysize (cell_size,cell_size) 
                        background Transform (Solid("#222"), size =  (cell_size,cell_size))

                        if cell:
                            for item in cell.items():                               
                                #text "[cell.state()]"
                                add item.image():
                                    size (cell_size,cell_size)


label test_display_field:
    $ obj = tField(10,15)
    $ obj.reset()        
    show screen display_field(obj)
    label .loop:
        pause   
    jump .loop 


        





        