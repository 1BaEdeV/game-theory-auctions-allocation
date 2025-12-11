import networkx as nx
import numpy as np

class DataLoader: 
    def load_karate_club(self) -> nx.Graph:
        
        return nx.karate_club_graph()

    def load_dolphins(self) -> nx.Graph:
        
        url = "http://www-personal.umich.edu/~mejn/netdata/dolphins.gml"
        return  nx.read_gml(url)
    

    def load_gml(self, url=None) -> nx.Graph:

        if url is None:
            raise ValueError("url не указан!")
        
        return nx.read_gml(url)