$$
\begin{align}
  \text{Prog} &\to [\text{Stmt}]^* \\
  [\text{Stmt}] &\to
  \begin{cases}
    \text{exit}([\text{Expr}]) \\
    \text{let} \ \text{ident} [: \text{type}] = [\text{Expr}] \\
    \text{const} \ \text{ident} [: \text{type}] = [\text{Expr}] \\
    \text{ident} = [\text{Expr}]
  \end{cases} \\
  [\text{Expr}] &\to [\text{Term}] ([+|-] [\text{Term}])^* \\
  [\text{Term}] &\to [\text{Factor}] ([*|/] [\text{Factor}])^* \\
  [\text{Factor}] &\to \text{int\_lit} \mid \text{ident} \mid ([\text{Expr}])
\end{align}
$$
