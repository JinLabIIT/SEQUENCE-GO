#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Created on Sat Feb  8 13:16:52 2020

@author: cheung
"""
import networkx as nx
from networkx.generators import random_graphs
import json
n = 64
d = 2
p = d/(n-1)
G = random_graphs.fast_gnp_random_graph(n,p,directed = True)
nx.drawing.nx_pylab.draw(G)
path = "1.json"
#nx.write_graph(G,path)
nx.drawing.nx_pylab.draw(G)

data = nx.readwrite.json_graph.node_link_data(G)
with open(path, 'w') as f:
    json.dump(data, f)