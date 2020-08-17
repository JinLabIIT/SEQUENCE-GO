#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Created on Sat Feb  8 13:16:52 2020

@author: cheung
"""
import networkx as nx
from networkx.generators import random_graphs
import json

for n in range(48, 81, 8):
    for d in range(3,6):
        d /= 2.0
        for seed in range(3):
            p = d/(n-1)
            for seed in range(3):
                G = random_graphs.fast_gnp_random_graph(n,p,seed=seed, directed = True)
                nx.drawing.nx_pylab.draw(G)
                path = "%d_%.1f_%d.json" % (n, d, seed)
                #nx.write_graph(G,path)
                nx.drawing.nx_pylab.draw(G)

                data = nx.readwrite.json_graph.node_link_data(G)
                with open(path, 'w') as f:
                    json.dump(data, f)
