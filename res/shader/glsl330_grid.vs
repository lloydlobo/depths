// https://www.vertexshaderart.com/src/#s=XQAAAQCzBAAAAAAAAAA9iIpmlGmcB7Xtn81bJpWuqJzDtplZfO9dezAKUzAhOq1A6%2FcsnKoBYBwJk9n5rbOkVWrNgjUljMM5MCmjtWFt1OTzDh5ZR63Eb0FXsAahCa%2FG44dntNxcwP%2BJXGe3nIdjgfDqQ7yh6imGJu4GVcp50%2F2LFOztDjVLUandNLOgHuhiMAnZ0DgL9J7Va4Oh9VqCcBTfvQ%2BtXqlOI0b16MzL0JowniFaA4U%2F4pkSvE%2BIA4A9nHYmpVH8psQlCnmQfwwKOSVPP4%2FZ88o56vDysjz3jbImdh4FjCO%2FeGCiB%2B3mD5QbNwOUhn%2FqLlvkk2Yv%2BvCPPvlZvcm1iWp7Q8CvOifdutCXACmGEKuch8dTgxkSktVAVqpzUXyNnhxA%2Fo8KzUVQioJJcLMO93MlfsLBVoQ%2FysThoC%2FztzJ7KYtB8I2zbw7Hj0k59B%2BUha7FTZ9VC7%2BHLjhbE57Nn%2FebfGrpN26udH5gv94GiRyBUL2PwFGLR1IB2aZBayfV0ohcrtp2np%2B717GZ6S4wqoUMoi4wKwyPF6EEq87DIqXYj7rM2jqBwSuKAX8iFJnFOAoGKKrGmRNWy7cQN1H9Kkc23fH4P00Dyqsd0GoI9FYPYFSxCaVVDA2olk65S0hJjmarPQcG8Zt%2F33a1ngNbibk9pxb8Vv2yGjdpK4vMGEI1xCfhUTKQ2PfegznuT%2B%2FkRHh6sjWF3izaEA1XKEMN1l75rF2exEY7n%2FGyVv%2Bkd8DFcZk3SF2UfB2UHAeQdOY5n9%2FOQL76PKNMJfWirwhpPb19rykgOMD%2FyQKtyg%3D%3D
void main() {

  float down = floor(sqrt(vertexCount)); // Check in help menu
  float across = floor(vertexCount/down); // 31.6; // sqrt(1000)->31 [1000**0.5]==31.6

  float x = mod(vertexId, across); // 0 1 2 3 4 5 ... 10 11 12
  //                                  0 1 2 3          0  1  2

  float y = floor(vertexId / across); // 0 1 2 3 4 5 ... 10 11 12
  //                                     0 0 0 0          1  1  1

  float u = x / ( across - 1.);
  float v = y / ( across - 1.);

  float ux = u*2.-1.; // 0 -> 1 => reverse normalize (sine like) -1 -> 1
  float vy = v*2.-1.;

  gl_Position = vec4(x, y, 0, 1); // two/three concetrated at the edges
  gl_Position = vec4(u, v, 0, 1); // concentrated grid of blocks
  gl_Position = vec4(ux, vy, 0, 1);

  gl_PointSize = 10.;
  gl_PointSize *= 20. / across; // Automatically adjust for size. Make them smaller (as 31 across)
  gl_PointSize *= resolution.x / 600.; // 1200/600 (get twicw as bit on changing screen size)

  v_color = vec4(1, 0, 0, 1);
}
