# Groupie-tracker

Régler le probléme de doublon 
Aprés idetification de la zone problématique par élimination, je vais utiliser le concept de Map[], une sorte de liste qui vas vérifier et eviter les doublons d'affichage
 - filteredMap := make(map[int]Relation)
 -  if _, exists := filteredMap[relation.ID]; !exists {
                    filteredMap[relation.ID] = relation

                    Cela nous évite la duplication, probléme réglé 

17/01 20h35 : Ajout des images correspondantes à chaque groupe, aprés avoir crée une variable pour recupérer le nom du groupe et l'inserer dynamiquement dans la template pour avoir le bon lien, suppression des espaces...

Résoudre les soucis de mise en forme du texte pour les images