# up - le plombier ultime

** up ** est le ** Ultimate Plumber **, un outil pour écrire des pipes Linux dans un
interface utilisateur interactive basée sur le terminal, avec aperçu instantané en direct des résultats de la commande.

L’objectif ** principal ** du plombier final est d’aider ** de manière interactive et
explorer de manière incrémentielle les données textuelles ** sous Linux, en facilitant la
construire des pipelines complexes, grâce à une ** boucle de rétroaction rapide **. Ceci est réalisé
en renforçant tous les ** utilitaires Linux ** de traitement de texte tels que `grep`,` sort`,
`cut`,` paste`, `awk`,` wc`, `perl`, etc., etc., en fournissant un rapide,
** aperçu interactif, défilable ** de leurs résultats.

[! [] (up.gif)] (https://asciinema.org/a/208538)

## Utilisation

** [Télécharger * en haut * pour Linux] (https://github.com/akavel/up/releases/download/v0.3/up) **
& nbsp; | & nbsp; [Autres systèmes d'exploitation] (https://github.com/akavel/up/releases)

Pour commencer à utiliser ** up **, redirigez-y toute commande (ou pipeline) émettant du texte.
- par exemple:

    $ lshw | & ./up

puis:

- utilisez *** PgUp / PgDn *** et *** Ctrl- [←] / Ctrl- [→] *** pour une navigation de base
  la sortie de la commande;
- dans la zone de saisie en haut de l'écran, commencez ** par écrire n'importe quel bash
  pipeline**; puis ** appuyez sur Entrée pour exécuter la commande que vous avez saisie **,
  et le plombier ultime vous montrera immédiatement la sortie de
  le pipeline dans la ** fenêtre à défilement ** ci-dessous (en remplacement de
  contenu précédent)
    - Par exemple, vous pouvez essayer d’écrire:
      `grep network -A2 | grep: | cut -d: -f2- | coller - -`
      - sur mon ordinateur, après avoir appuyé sur * Entrée *, l'écran affiche
      le pipeline et un aperçu déroulant de sa sortie comme ci-dessous:

             | réseau grep -A2 | grep: | cut -d: -f2- | coller - -
             Interface sans fil Centrino Advanced-N 6235
             Interface Ethernet Contrôleur Ethernet Gigabit PCI Express RTL8111 / 8168/8411

    - ** AVERTISSEMENT: soyez prudent lorsque vous l'utilisez! Cela pourrait être dangereux. **
      En particulier, écrire "rm" ou "dd" dedans pourrait ressembler à courir
      avec une scie à chaîne. Mais vous feriez bien d'écrire "rm" n'importe où dans Linux
      de toute façon, non?
- lorsque vous êtes satisfait du résultat, vous pouvez ** appuyer sur * Ctrl-X * pour quitter **
  le plombier ultime, et la commande que vous avez construite sera ** écrite dans
  `up1.sh` fichier ** dans le répertoire de travail actuel (ou, s'il existait déjà,
  `up2.sh`, etc., jusqu’à 1000, sur la base de [Shlemiel le peintre
  algorithme] (https://www.joelonsoftware.com/2001/12/11/back-to-basics/)).
  Vous pouvez également appuyer sur *** Ctrl-C *** pour quitter sans enregistrer.
- Si la commande dans laquelle vous transmettez * up * dure longtemps (dans ce cas, vous verrez
  un tilde `~` caractère indicateur dans le coin supérieur gauche de l'écran, ce qui signifie
  que * up * attend toujours plus d’entrée), vous devrez peut-être appuyer sur
  *** Ctrl-S *** pour geler temporairement le tampon d'entrée de * up * (un gel sera
  indiqué par un caractère «#» dans le coin supérieur gauche), qui injectera un faux
  EOF dans le pipeline; sinon, certaines commandes du pipeline risquent de ne pas être imprimées
  quoi que ce soit, en attente d'une entrée complète (en particulier des commandes comme `wc` ou` sort`,
  mais `grep`,` perl`, etc. peuvent également donner des résultats incomplets). Dégeler,
  appuyez sur *** Ctrl-Q ***.
## Notes complémentaires

- Le pipeline est transmis textuellement à une commande `bash -c`, ainsi tout bash-isms devrait fonctionner.
- La mémoire tampon d'entrée de Ultimate Plumber est actuellement fixée à ** 40 Mo **. Si
  vous atteignez cette limite, un caractère `+` devrait s'afficher en haut à gauche
  coin de l'écran. (Ceci est destiné à être changé en un
  développable de manière dynamique / manuelle dans une future version de * up *.)
- ** Prise en charge de MacOSX: ** Je n'ai pas de Mac, donc je ne sais pas s'il fonctionne sur
  un. Vous êtes invités à essayer et à envoyer des relations publiques. Si vous êtes intéressé par
  me fournissant une sorte de support de type officiel pour MacOSX, veuillez considérer
  essayant de trouver un moyen de m'envoyer un ordinateur Mac utilisable. Notez s'il vous plaît
  Je n'essaie pas de "profiter" de cela, car je ne suis en fait pas du tout
  intéressé à atteindre un Mac autrement. (En outre, en essayant de s'engager à ce genre
  de soutien sera un fardeau supplémentaire et une obligation pour moi. Connaître quelqu'un
  il se soucie assez de faire un geste physique de fantaisie aiderait vraiment à soulager
  Si vous êtes suffisamment sérieux pour envisager cette option, veuillez me contacter par
  email (mailto: czapkofan@gmail.com) ou keybase (https://keybase.io/akavel), donc
  que nous pourrions essayer de rechercher des moyens pour y parvenir.
  Merci de votre compréhension!
- ** Etat de la technique: ** J'étais surpris que personne ne semble avoir écrit un outil similaire auparavant,
  que j'ai pu trouver. Il aurait dû être possible d'écrire cela depuis l'aube
  de Unix déjà, ou plus tôt! Et en effet, après avoir annoncé * up *, j'en ai assez
  annonce que mon attention a déjà été portée sur un de ces projets précédents:
  ** [Pipecut] (http://pipecut.org/index.html) **. Semble intéressant! Tu peux aimer
  pour le vérifier aussi! (Merci [@TronDD] (https://lobste.rs/s/acpz00/up_tool_for_writing_linux_pipes_with#c_qxrgoa).)
- ** Autres influences: ** Je ne me rappelle pas trop bien le fait, mais je suis
  Plutôt sûr que cela doit avoir été inspiré en grande partie par The Bret Victor's Talk.

## Idées futures

- J'ai pas mal d'idées pour poursuivre l'expérimentation du développement de
  * up *, y compris mais sans s'y limiter:
    - [RIIR] (https://rust-lang.org) (une fois que j'en saurai assez sur Rust ... chez certains
      point à l'avenir ... peut-être ...) - esp. j'espère que maquille * être * un plus petit
      binaire (et peut-être enfin apprendre un peu de rouille); si je suis un peu
      peur si cela pourrait ossifier la base de code et rendre plus difficile à développer
      plus loin..? ... mais peut-être réellement converser? ...
    - Peut-être que cela pourrait être transformé en une interface sans interface utilisateur, RPC / REST / socket / text-driven
      comme gocode ou [Language Servers] (https://langserver.org/), par exemple,
      intégration avec les éditeurs / IDE (emacs? vim? VSCode? ...) Je serais particulièrement
      intéressé à le fusionner éventuellement dans [Luna
      Studio] (https://luna-lang.org/); Le RIIR peut aider à cela. (Avant cela, comme
      une approche plus simple, une édition sur plusieurs lignes peut être nécessaire, ou du moins
      défilement gauche et droit de la zone de saisie de l'éditeur de commandes. En outre, une sorte de
      sauter entre les mots dans la ligne de commande; les lignes de lecture * Alt-b * et * Alt-f *?)
    - Permettre de [capturer la sortie de déjà en cours
      processus] (https://stackoverflow.com/a/19584979/98528)! (Mais peut-être que
      pourrait être mieux fait comme un outil distinct, composable! En rouille?)
    - Ajout de tests ... (ahem; voir aussi
      [# 1] (https://github.com/akavel/up/issues/1)) ... écrivez aussi `--help` ...
    - Le faire fonctionner sous Windows,
      en quelque sorte [?] (https://github.com/mattn/go-shellwords) Aussi, évidemment,
      être agréable d'avoir une infrastructure CI permettant de le porter sur MacOSX,
      BSD, etc., etc ...
    - Intégration avec [fzf] (https://github.com/junegunn/fzf) et d'autres TUI
      outils? Je n'ai que quelques idées et idées vagues à ce sujet à partir de maintenant, pas
      même sûr à quoi cela pourrait ressembler.
    - Ajout de plus de prévisualisations, pour chaque `|` du pipeline; aussi forking de
      pipelines, fusion, boucles de rétroaction et autres opérations de mélange et d’appariement (bien que
      Je préférerais fortement que [Luna] (https://luna-lang.org) le fasse
      finalement).
- Si vous souhaitez financer mes travaux de R & D, contactez-moi par courrier électronique à l'adresse suivante:
  czapkofan@gmail.com, ou [sur keybase.io en tant que akavel] (https://keybase.io/akavel).
  Je suppose que je développerai probablement encore le Ultimate Plumber,
  mais pour le moment, c’est purement un projet de loisir, avec tout le plaisir et les risques que cela comporte.
  implique.

- * Mateusz Czapliński *
* Octobre 2018 *
