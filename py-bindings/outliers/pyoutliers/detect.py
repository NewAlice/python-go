from __future__ import annotations  # min Py version 3.7
import numpy as np


def detect(data: list[float]) -> list[int]:
    """Return indices where values more than 2 standard deviations from mean"""
    data = np.fromiter(data, dtype='float64')  # data: np.ndarray[np.float64]
    indices = np.where(np.abs(data - data.mean()) > 2 * data.std())[0]
    return(indices.tolist())  # return: list[int]


def gen_testdata() -> list[float]:
    """Return testdata"""
    data = np.random.rand(1000)
    indices = [7, 113, 835]
    for i in indices:
        data[i] += 97

    return(data.tolist())
