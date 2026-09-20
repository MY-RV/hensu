# Por qué existe Hensu

Le preguntás a un asistente si la contraseña de la base ya está puesta. Hace lo obvio: abre el `.env`. Ahora esa contraseña está en el chat, en el log de la sesión y en el contexto de un modelo. Tres lugares de los que no se va.

Nadie hizo nada mal. La pregunta era chica — ¿está puesta o no? — y la respuesta vino entera.

Hensu no nació como producto. Salió del día a día con **Naoki** en [Rosvelt](https://rosvelt.com/): un script adentro del repo, `scripts/service/dotenv`, que fue creciendo porque todo terminaba necesitando lo mismo — agarrar la configuración del servicio, meterla en el entorno y correr el programa. Levantar el servicio. Sembrar la base. Correr los tests. Abrir el workbench. Seis cosas distintas haciendo el mismo baile, cada una con su propia copia del baile.

Ese script ya sabía tapar valores: podía imprimir `***` en vez del secreto. Pero tapar era opcional. Había que elegir el verbo cuidadoso, y eso funciona mientras el único que lee es alguien que sabe lo que hace.

Cambiaron los lectores. Un modelo al que le pedís que verifique una key no es descuidado — hace lo obvio, y abrir el archivo *es* lo obvio. Pedirle que se acuerde del comando prudente es pedirle que tenga una costumbre. Una costumbre no es un mecanismo.

Así que se dio vuelta el default. Ahora un valor llega recortado siempre, salvo que alguien escriba la palabra:

```bash
hensu get PORT            # {"PORT": {"value": "***", "defined": true}}
hensu -r trust get PORT   # {"PORT": {"value": "8080", "defined": true}}
```

Y cuando el valor hace falta de verdad, casi nunca lo necesita quien pregunta: lo necesita un programa. Entonces se lo damos al programa y ya. `hensu exec -- go run .` corre tu app con la configuración adentro, y vos no viste nada.

Esto no es una jaula, y conviene decirlo fuerte. Cualquiera con una terminal hace `cat .env` y lo ve todo. No hay CLI que lo impida. Lo que cambió es otra cosa: el camino cómodo dejó de ser el que filtra, y una fuga pasó a ser algo que alguien escribió a propósito, en un comando que mañana podés ir a buscar.

Es la segunda CLI publicada de la familia: primero [GoDo](https://github.com/MY-RV/godo), y antes de todo [goxdi](https://github.com/MY-RV/goxdi). godo no quedó de adorno — es el que corre la puerta de calidad de Hensu, `godo ci` en verde antes de cada tag.

El nombre es **変数** — *hensū*, variable. 変 es cambio, 数 es valor. Eso es un archivo de configuración: valores que cambian según la máquina, con nombre para que un programa los encuentre.

Si confiás en todos los que están en la sala, no necesitás nada de esto. Yo dejé de estar seguro de quién está en la sala.

[Getting started](./getting-started.md) · [English](./story.md)
