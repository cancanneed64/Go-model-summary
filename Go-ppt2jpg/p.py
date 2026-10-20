# -*- codeing = utf-8 -*-
# @Time : 2022/5/1 13:23
# @Author : ning
# @File : 5.1demo.py
# @software : PyCharm

import matplotlib.pyplot as plt
from PIL import Image
import comtypes.client
import os
import win32
import glob



#逻辑
#判断源文件格式，如果是ppt，转换为jpg再调整格式
#如果是png，转换为jpg再调整格式
#如果是jpg，直接调整格式

#png转为jpg
def pngtojpg(name):
    #如果是png，转换为jpg再调整格式
    img = Image.open(name)
    if img.mode == "RGBA":
        img = img.convert('RGB')
        print(name)
        imgname = os.path.splitext(name)[0]+ ".jpg"
        print(imgname)
        img.save(imgname)

        
#ppt转为jgp
def ppttojpg(openpath):

    #fileNames = glob.glob(openpath + r'\*')   
##    for fileName in fileNames:     #将pa 文件夹中的文件删除。
##        os.remove( fileName)
##    powerpoint = comtypes.client.CreateObject("kwpp.Application") #使用wps的接口
    powerpoint = comtypes.client.CreateObject("Powerpoint.Application") 
    #powerpoint.Visible = 0
    ppt = powerpoint.Presentations.Open(openpath,WithWindow=0)
    # 另存为
    pa = 'C:\\Users\\byao394\\Desktop\\f'
    ppt.SaveAs(pa + '.jpg', 17)
    # 退出
    ppt.Close()
    powerpoint.Quit()


def adjust_image(file_in, width, height, file_out):
    
    image = Image.open(file_in)
    resized_image = image.resize((width, height), Image.ANTIALIAS)
    resized_image.save(file_out)


if __name__ == '__main__':


    
    
    width = 1920 #调整的分辨率大小
    height = 1080
    n = 0

    #设定路径
    #源文件目录
    my_dir = 'C:\\Users\\byao394\\Desktop\\f'
    Files = os.listdir(my_dir)
    print(os.getcwd())
    
    #遍历路径下的文件
    for k in range(len(Files)):

        #将文件名
        #Files[k]=os.path.splitext(Files[k])[1]
        print(Files[k])
        file_in = ""
        if Files[k].endswith('py'):
            continue
        
        elif Files[k].endswith('jpg'):
            
            file_in = Files[k]
        elif Files[k].endswith('png'):
            pngtojpg(Files[k])
            file_in = os.path.splitext(Files[k])[0]+ ".jpg"
            
        elif Files[k].endswith('pptx'):
            n = n+1
            #cur_dir = os.path.dirname(os.path.abspath(Files[k]))
            #path1 = os.path.join(os.path.abspath(my_dir + os.path.sep + ".."), Files[k])
            #print(path1)
            input_file_path = os.path.join(my_dir, Files[k])
            print(input_file_path)
            ppttojpg(input_file_path)
            file_in = "幻灯片{0}.JPG".format(n)
            
        
        file_out = 'resized'+ file_in
        # 调整分辨率
        adjust_image(file_in, width, height, file_out)

        
##    Str2=['.pptx','.png','.jpg']
##    print(set(Str2).intersection(set(Files)))
##
##    if len(list(set(Str2).intersection(set(Files))))==len(Str2):
##        print("yes")
##        #return True
##    else:
##        print("no")
##        #return False


